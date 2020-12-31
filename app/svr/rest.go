// Package svr provides the application server implementations.
package svr

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/bodenr/vehicle-api/config"
	"github.com/bodenr/vehicle-api/svr/proto"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/hlog"

	"github.com/bodenr/vehicle-api/log"
	"github.com/bodenr/vehicle-api/util"

	protobuf "github.com/golang/protobuf/proto"
)

// StoreError is a store specific error that can also include a status code hint.
type StoreError struct {
	// Error is a reference to the actual error.
	Error error

	// StatusCode is a hint of the type of error incurred; based on http status codes.
	StatusCode int
}

// RestServer wraps a http.Server reference.
type RestServer struct {
	*http.Server
}

// RequestVars is a map of string to string values representing the mux.Vars for a request.
type RequestVars map[string]string

// StoredResource encapsulates the logic for a API accessible resource.
type StoredResource interface {
	// CreateSchema creates the datastore schema for the resource.
	CreateSchema()

	// Search for resources given the said query values.
	Search(url.Values) ([]interface{}, *StoreError)

	// List all stored resources.
	List() ([]interface{}, *StoreError)

	// Get a single stored resource based on the request vars.
	Get(RequestVars) (interface{}, *StoreError)

	// Delete a single stored resource based on the request vars.
	Delete(RequestVars) *StoreError

	// Create a new stored resource.
	Create(interface{}) (interface{}, *StoreError)

	// Update an existing stored resource.
	Update(interface{}, RequestVars) (interface{}, *StoreError)

	// GetETag returns the eTag for a single resource based on request vars.
	GetETag(RequestVars) (string, *StoreError)

	// BuildETag returns the eTag for the said stored resource.
	BuildETag(interface{}) (string, error)

	// Unmarshal the byte slice with the said content type.
	Unmarshal(contentType string, resource []byte) (interface{}, error)

	// Marshal the said resource into the said content type.
	Marshal(contentType string, resource interface{}) ([]byte, error)

	// Validate the said resource prior to create/update.
	Validate(resource interface{}, httpMethod string) error

	// Bind the stored resource routes into the router.
	BindRoutes(router *mux.Router)
}

// RestfulResource wraps a StoredResource.
type RestfulResource struct {
	Resource StoredResource
}

// RestfulHandler provides the methods supporting REST API handling for a StoredResource.
type RestfulHandler interface {
	// TODO: support partial updates via PATCH

	// Create handles REST API create requests for a StoredResource.
	Create(writer http.ResponseWriter, request *http.Request)

	// Create handles REST API list requests for a StoredResource.
	List(writer http.ResponseWriter, request *http.Request)

	// Create handles REST API delete requests for a StoredResource.
	Delete(writer http.ResponseWriter, request *http.Request)

	// Create handles REST API get requests for a StoredResource.
	Get(writer http.ResponseWriter, request *http.Request)

	// Create handles REST API update requests for a StoredResource.
	Update(writer http.ResponseWriter, request *http.Request)

	// Respond to the request with the given code and optional payload.
	Respond(writer http.ResponseWriter, request *http.Request, code int, payload interface{})

	// Respond to the request with an ErrorResponse.
	RespondErr(writer http.ResponseWriter, request *http.Request,
		code int, respErr proto.ErrorResponse)

	// Respond to the request with the given etag, code, and optional payload.
	RespondETag(writer http.ResponseWriter,
		request *http.Request, code int, payload interface{}, etag string)
}

// ETagExpires is a time in the distant past used on Expires header to disable time based caching.
var ETagExpires = util.TimeFromMillis(0).String()

// NewRestfulResource creates a new RestfulResource for the said StoredResource
func NewRestfulResource(storedResource StoredResource) RestfulResource {
	return RestfulResource{
		Resource: storedResource,
	}
}

// Create handles the REST API logic to create its underlying StoredResource.
func (handler RestfulResource) Create(writer http.ResponseWriter, request *http.Request) {
	// TODO: enforce max size
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Log.Err(err).Msg("Error ready request body")
		handler.Respond(writer, request, http.StatusBadRequest, nil)
		return
	}
	contentType := GetRequestContentType(request)
	if contentType == "" {
		handler.Respond(writer, request, http.StatusUnsupportedMediaType, nil)
		return
	}
	// TODO: better validation/sanitization
	resource, err := handler.Resource.Unmarshal(contentType, body)
	if err != nil {
		log.Log.Err(err).Msg("Invalid resource request body")
		handler.RespondErr(writer, request, http.StatusBadRequest, proto.ErrorResponse{Message: err.Error()})
		return
	}
	if err = handler.Resource.Validate(resource, request.Method); err != nil {
		log.Log.Err(err).Msg("Invalid resource format")
		handler.RespondErr(writer, request, http.StatusBadRequest, proto.ErrorResponse{Message: err.Error()})
		return
	}

	resource, sErr := handler.Resource.Create(resource)
	if sErr != nil {
		handler.RespondErr(writer, request, sErr.StatusCode, proto.ErrorResponse{Message: sErr.Error.Error()})
		return
	}

	etag, err := handler.Resource.BuildETag(resource)
	if err != nil {
		log.Log.Err(err).Msg("Failed to build etag")
		handler.Respond(writer, request, http.StatusOK, resource)
		return
	}
	handler.RespondETag(writer, request, http.StatusOK, resource, etag)
}

// Create handles the REST API logic to list its underlying StoredResources.
func (handler RestfulResource) List(writer http.ResponseWriter, request *http.Request) {
	var err *StoreError
	var resources []interface{}

	queryParams := request.URL.Query()
	if len(queryParams) == 0 {
		resources, err = handler.Resource.List()
	} else {
		resources, err = handler.Resource.Search(queryParams)
	}

	if err != nil {
		handler.RespondErr(writer, request, err.StatusCode,
			proto.ErrorResponse{Message: err.Error.Error()})
		return
	}
	handler.Respond(writer, request, http.StatusOK, resources)
}

// Create handles the REST API logic to delete its underlying StoredResource.
func (handler RestfulResource) Delete(writer http.ResponseWriter, request *http.Request) {
	// TODO: collapse etag handling into shared method
	requestVars := mux.Vars(request)
	reqETag := request.Header.Get("If-None-Match")
	if reqETag != "" {
		resourceETag, sErr := handler.Resource.GetETag(requestVars)
		if sErr != nil {
			if sErr.StatusCode == http.StatusNotFound {
				handler.Respond(writer, request, http.StatusNotFound, nil)
				return
			}
			handler.RespondErr(writer, request, sErr.StatusCode, proto.ErrorResponse{Message: sErr.Error.Error()})
			return
		}
		if reqETag != resourceETag {
			handler.Respond(writer, request, http.StatusPreconditionFailed, nil)
			return
		}
	}

	err := handler.Resource.Delete(mux.Vars(request))
	if err != nil {
		if err.StatusCode == http.StatusNotFound {
			handler.Respond(writer, request, err.StatusCode, nil)
			return
		}
		handler.RespondErr(writer, request, err.StatusCode, proto.ErrorResponse{Message: err.Error.Error()})
		return
	}
	handler.Respond(writer, request, http.StatusNoContent, nil)
}

// Create handles the REST API logic to get a specific underlying StoredResource.
func (handler RestfulResource) Get(writer http.ResponseWriter, request *http.Request) {
	requestVars := mux.Vars(request)
	reqETag := request.Header.Get("If-None-Match")
	if reqETag != "" {
		resourceETag, sErr := handler.Resource.GetETag(requestVars)
		if sErr != nil {
			if sErr.StatusCode == http.StatusNotFound {
				handler.Respond(writer, request, http.StatusNotFound, nil)
				return
			}
			handler.RespondErr(writer, request, sErr.StatusCode, proto.ErrorResponse{Message: sErr.Error.Error()})
			return
		}
		if reqETag == resourceETag {
			handler.Respond(writer, request, http.StatusNotModified, nil)
			return
		}
	}

	resource, sErr := handler.Resource.Get(mux.Vars(request))
	if sErr != nil {
		if sErr.StatusCode == http.StatusNotFound {
			handler.Respond(writer, request, sErr.StatusCode, nil)
			return
		}
		handler.RespondErr(writer, request, sErr.StatusCode, proto.ErrorResponse{Message: sErr.Error.Error()})
		return
	}

	etag, err := handler.Resource.BuildETag(resource)
	if err != nil {
		log.Log.Err(err).Msg("Failed to build etag")
		handler.Respond(writer, request, http.StatusOK, resource)
		return
	}
	handler.RespondETag(writer, request, http.StatusOK, resource, etag)
}

// Create handles the REST API logic to update a specific underlying StoredResource.
func (handler RestfulResource) Update(writer http.ResponseWriter, request *http.Request) {
	requestVars := mux.Vars(request)
	reqETag := request.Header.Get("If-None-Match")
	if reqETag != "" {
		resourceETag, sErr := handler.Resource.GetETag(requestVars)
		if sErr != nil {
			if sErr.StatusCode == http.StatusNotFound {
				handler.Respond(writer, request, http.StatusNotFound, nil)
				return
			}
			handler.RespondErr(writer, request, sErr.StatusCode, proto.ErrorResponse{Message: sErr.Error.Error()})
			return
		}
		if reqETag != resourceETag {
			handler.Respond(writer, request, http.StatusPreconditionFailed, nil)
			return
		}
	}

	// TODO: enforce max size
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Log.Err(err).Msg("Error reading request body")
		handler.Respond(writer, request, http.StatusInternalServerError, nil)
		return
	}
	contentType := GetRequestContentType(request)
	if contentType == "" {
		handler.Respond(writer, request, http.StatusUnsupportedMediaType, nil)
		return
	}
	// TODO: better validation
	resource, err := handler.Resource.Unmarshal(contentType, body)
	if err != nil {
		log.Log.Err(err).Msg("Error unmarshalling request body")
		handler.RespondErr(writer, request, http.StatusBadRequest, proto.ErrorResponse{Message: err.Error()})
		return
	}
	if err = handler.Resource.Validate(resource, request.Method); err != nil {
		log.Log.Err(err).Msg("Invalid body format")
		handler.RespondErr(writer, request, http.StatusBadRequest, proto.ErrorResponse{Message: err.Error()})
		return
	}
	resource, sErr := handler.Resource.Update(resource, mux.Vars(request))
	if sErr != nil {
		handler.RespondErr(writer, request, sErr.StatusCode, proto.ErrorResponse{Message: sErr.Error.Error()})
		return
	}
	etag, err := handler.Resource.BuildETag(resource)
	if err != nil {
		log.Log.Err(err).Msg("Failed to build etag")
		handler.Respond(writer, request, http.StatusOK, resource)
		return
	}
	handler.RespondETag(writer, request, http.StatusOK, resource, etag)
}

// RespondETag responds to the http request with eTag based headers.
func (handler RestfulResource) RespondETag(writer http.ResponseWriter,
	request *http.Request, code int, payload interface{}, etag string) {

	writer.Header().Set("Pragma", "no-cache")
	writer.Header().Set("Expires", ETagExpires)
	writer.Header().Set("ETag", etag)

	handler.Respond(writer, request, code, payload)
}

// RespondErr to the request with an ErrorResponse.
func (handler RestfulResource) RespondErr(writer http.ResponseWriter, request *http.Request,
	code int, respErr proto.ErrorResponse) {

	// TODO: special handling for protobuf error marshalling sux
	if GetResponseContentType(request) == ContentAppProtobuf {
		responseBody, err := protobuf.Marshal(&respErr)
		if err != nil {
			log.Log.Err(err).Msg("Error marshalling protobuf")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", ContentAppProtobuf)
		writer.WriteHeader(code)
		writer.Write(responseBody)
	} else {
		handler.Respond(writer, request, code, respErr)
	}
}

// Respond to the http request with the said status code and optional payload.
func (handler RestfulResource) Respond(writer http.ResponseWriter,
	request *http.Request, code int, payload interface{}) {

	if payload != nil {
		contentType := GetResponseContentType(request)
		responseBody, err := handler.Resource.Marshal(contentType, payload)
		if err != nil {
			responseBody, _ = Marshal(contentType, proto.ErrorResponse{Message: "Error marshalling response body"})
			code = http.StatusInternalServerError
		}
		writer.Header().Set("Content-Type", contentType)
		writer.WriteHeader(code)
		writer.Write(responseBody)
	} else {
		writer.WriteHeader(code)
	}
}

// NewRestServer creates a new RestServer for the given config that will expose the given StoredResources.
func NewRestServer(conf *config.HTTPConfig, storedResources ...StoredResource) *RestServer {

	router := mux.NewRouter()

	subrouter := router.PathPrefix("/api").Subrouter()

	subrouter.Use(hlog.NewHandler(log.Log))

	subrouter.Use(hlog.AccessHandler(
		func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("")
		}))

	subrouter.Use(hlog.RemoteAddrHandler("ip"))
	subrouter.Use(hlog.UserAgentHandler("user_agent"))
	subrouter.Use(hlog.RefererHandler("referer"))
	subrouter.Use(hlog.RequestIDHandler("req_id", "Request-Id"))

	for _, resource := range storedResources {
		resource.BindRoutes(subrouter)
	}

	return &RestServer{
		Server: &http.Server{
			Addr:         conf.Address,
			Handler:      subrouter,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Run starts the RestServer and is blocking in nature.
func (server *RestServer) Run() error {
	log.Log.Info().Msg("starting http server on port " + server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Log.Err(err).Msg("failed to start http server")
		return err
	}
	return nil
}

// WaitForShutdown waits for a signal on terminate and then shutsdown the RestServer.
func (server *RestServer) WaitForShutdown(terminate <-chan os.Signal, done chan bool) {
	termSig := <-terminate

	log.Log.Info().Str(log.Signal, termSig.String()).Msg("shutting down http server")

	ctx, halt := context.WithTimeout(context.Background(), 20*time.Second)
	defer halt()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Log.Err(err).Msg("failed to gracefully stop http server")
	} else {
		log.Log.Info().Msg("graceful shutdown of http server complete")
	}

	close(done)
}
