package svr

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bodenr/vehicle-api/config"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/hlog"

	"github.com/bodenr/vehicle-api/log"
)

type Error struct {
	Message string `json:"error_message,omitempty"`
}

type StoreError struct {
	Error      error
	StatusCode int
}

type RestServer struct {
	*http.Server
}

type RequestVars map[string]string

type StoredResource interface {
	Search(queryParams *url.Values) ([]StoredResource, *StoreError)
	List() ([]StoredResource, *StoreError)
	Get(requestVars RequestVars) (StoredResource, *StoreError)
	Delete(requestVars RequestVars) *StoreError
	Create() *StoreError
	Update(requestVars RequestVars) *StoreError
	Unmarshal(contentType string, resource []byte) (StoredResource, error)
	Validate(method string) error
	BindRoutes(router *mux.Router, handler RestfulHandler)
}

type RestfulResource struct {
	Resource StoredResource
}

type RestfulHandler interface {
	Create(writer http.ResponseWriter, request *http.Request)
	List(writer http.ResponseWriter, request *http.Request)
	Delete(writer http.ResponseWriter, request *http.Request)
	Get(writer http.ResponseWriter, request *http.Request)
	Update(writer http.ResponseWriter, request *http.Request)
}

func NewRestfulResource(storedResource StoredResource, router *mux.Router) RestfulResource {
	handler := RestfulResource{
		Resource: storedResource,
	}
	storedResource.BindRoutes(router, handler)
	return handler
}

func (handler RestfulResource) Create(writer http.ResponseWriter, request *http.Request) {
	// TODO: enforce max size
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Log.Err(err).Msg("Error ready request body")
		HttpRespond(writer, request, http.StatusBadRequest, nil)
		return
	}
	contentType := GetRequestContentType(request)
	if contentType == "" {
		HttpRespond(writer, request, http.StatusUnsupportedMediaType, nil)
		return
	}
	// TODO: better validation/sanitization
	resource, err := handler.Resource.Unmarshal(contentType, body)
	if err != nil {
		log.Log.Err(err).Msg("Invalid resource request body")
		HttpRespond(writer, request, http.StatusBadRequest, Error{Message: err.Error()})
		return
	}
	if err = resource.Validate(request.Method); err != nil {
		log.Log.Err(err).Msg("Invalid resource format")
		HttpRespond(writer, request, http.StatusBadRequest, Error{Message: err.Error()})
		return
	}

	if err := resource.Create(); err != nil {
		// TODO: ideally don't use error string parsing
		if strings.Contains(err.Error.Error(), "duplicate key") {
			HttpRespond(writer, request, http.StatusBadRequest,
				Error{Message: "Resource already exists"})
			return
		}
		HttpRespond(writer, request, http.StatusInternalServerError, nil)
		return
	}

	HttpRespond(writer, request, http.StatusOK, resource)
}

func (handler RestfulResource) List(writer http.ResponseWriter, request *http.Request) {
	var err *StoreError
	var resources []StoredResource

	queryParams := request.URL.Query()
	if len(queryParams) == 0 {
		resources, err = handler.Resource.List()
	} else {
		resources, err = handler.Resource.Search(&queryParams)
	}

	if err != nil {
		HttpRespond(writer, request, http.StatusInternalServerError,
			Error{Message: "Error listing resources"})
		return
	}
	HttpRespond(writer, request, http.StatusOK, resources)
}

func (handler RestfulResource) Delete(writer http.ResponseWriter, request *http.Request) {
	err := handler.Resource.Delete(mux.Vars(request))
	if err != nil {
		HttpRespond(writer, request, err.StatusCode, nil)
		return
	}
	HttpRespond(writer, request, http.StatusNoContent, nil)
}

func (handler RestfulResource) Get(writer http.ResponseWriter, request *http.Request) {
	resource, err := handler.Resource.Get(mux.Vars(request))
	if err != nil {
		HttpRespond(writer, request, err.StatusCode, nil)
		return
	}
	HttpRespond(writer, request, http.StatusOK, resource)
}

func (handler RestfulResource) Update(writer http.ResponseWriter, request *http.Request) {
	// TODO: enforce max size
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Log.Err(err).Msg("Error ready request body")
		HttpRespond(writer, request, http.StatusBadRequest, nil)
		return
	}
	contentType := GetRequestContentType(request)
	if contentType == "" {
		HttpRespond(writer, request, http.StatusUnsupportedMediaType, nil)
		return
	}
	// TODO: better validation
	resource, err := handler.Resource.Unmarshal(contentType, body)
	if err != nil {
		log.Log.Err(err).Msg("Error unmarshalling request body")
		HttpRespond(writer, request, http.StatusBadRequest, nil)
		return
	}
	if err = resource.Validate(request.Method); err != nil {
		log.Log.Err(err).Msg("Invalid body format")
		HttpRespond(writer, request, http.StatusBadRequest, Error{Message: err.Error()})
		return
	}
	if err := resource.Update(mux.Vars(request)); err != nil {
		HttpRespond(writer, request, err.StatusCode, Error{Message: err.Error.Error()})
		return
	}
	HttpRespond(writer, request, http.StatusOK, resource)
}

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
		NewRestfulResource(resource, subrouter)
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

func (server *RestServer) Run() error {
	log.Log.Info().Msg("starting http server on port " + server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Log.Err(err).Msg("failed to start http server")
		return err
	}
	return nil
}

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

func GetRequestContentType(request *http.Request) string {
	for _, contentType := range request.Header.Values("Content-Type") {
		if SupportsEncoding(contentType) {
			return contentType
		}
	}
	return ""
}

func GetResponseContentType(request *http.Request) string {
	// TODO: support weighted encodings, */*, etc..
	for _, acceptType := range request.Header.Values("Accept") {
		if SupportsEncoding(acceptType) {
			return acceptType
		}
	}
	// no Accept specified; if Content-Type was given use it
	contentType := GetRequestContentType(request)
	if contentType != "" {
		return contentType
	}
	// default to JSON
	return "application/json"
}

func HttpRespond(writer http.ResponseWriter, request *http.Request, code int, payload interface{}) {

	if payload != nil {
		contentType := GetResponseContentType(request)
		responseBody, err := Marshal(contentType, payload)
		if err != nil {
			responseBody, _ = Marshal(contentType, Error{Message: "Error marshalling response body"})
			code = http.StatusInternalServerError
		}
		writer.Header().Set("Content-Type", contentType)
		writer.WriteHeader(code)
		writer.Write(responseBody)
	} else {
		writer.WriteHeader(code)
	}

}
