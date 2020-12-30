package svr

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/bodenr/vehicle-api/config"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/hlog"

	"github.com/bodenr/vehicle-api/log"
)

type ErrorResponse struct {
	Message string `json:"error_message,omitempty" xml:"error_message"`
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
	CreateSchema()

	Search(url.Values) ([]interface{}, *StoreError)
	List() ([]interface{}, *StoreError)
	Get(RequestVars) (interface{}, *StoreError)
	Delete(RequestVars) *StoreError
	Create(interface{}) (interface{}, *StoreError)
	Update(interface{}, RequestVars) (interface{}, *StoreError)

	Unmarshal(contentType string, resource []byte) (interface{}, error)
	Marshal(contentType string, resource interface{}) ([]byte, error)
	Validate(resource interface{}, httpMethod string) error
	BindRoutes(router *mux.Router)
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

	HttpRespond(writer http.ResponseWriter, request *http.Request, code int, payload interface{})
}

func NewRestfulResource(storedResource StoredResource) RestfulResource {
	return RestfulResource{
		Resource: storedResource,
	}
}

func (handler RestfulResource) Create(writer http.ResponseWriter, request *http.Request) {
	// TODO: enforce max size
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Log.Err(err).Msg("Error ready request body")
		handler.HttpRespond(writer, request, http.StatusBadRequest, nil)
		return
	}
	contentType := GetRequestContentType(request)
	if contentType == "" {
		handler.HttpRespond(writer, request, http.StatusUnsupportedMediaType, nil)
		return
	}
	// TODO: better validation/sanitization
	resource, err := handler.Resource.Unmarshal(contentType, body)
	if err != nil {
		log.Log.Err(err).Msg("Invalid resource request body")
		handler.HttpRespond(writer, request, http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}
	if err = handler.Resource.Validate(resource, request.Method); err != nil {
		log.Log.Err(err).Msg("Invalid resource format")
		handler.HttpRespond(writer, request, http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}

	resource, sErr := handler.Resource.Create(resource)
	if sErr != nil {
		handler.HttpRespond(writer, request, sErr.StatusCode, ErrorResponse{Message: sErr.Error.Error()})
		return
	}

	handler.HttpRespond(writer, request, http.StatusOK, resource)
}

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
		handler.HttpRespond(writer, request, err.StatusCode,
			ErrorResponse{Message: err.Error.Error()})
		return
	}
	handler.HttpRespond(writer, request, http.StatusOK, resources)
}

func (handler RestfulResource) Delete(writer http.ResponseWriter, request *http.Request) {
	err := handler.Resource.Delete(mux.Vars(request))
	if err != nil {
		handler.HttpRespond(writer, request, err.StatusCode, ErrorResponse{Message: err.Error.Error()})
		return
	}
	handler.HttpRespond(writer, request, http.StatusNoContent, nil)
}

func (handler RestfulResource) Get(writer http.ResponseWriter, request *http.Request) {
	resource, err := handler.Resource.Get(mux.Vars(request))
	if err != nil {
		handler.HttpRespond(writer, request, err.StatusCode, ErrorResponse{Message: err.Error.Error()})
		return
	}
	handler.HttpRespond(writer, request, http.StatusOK, resource)
}

func (handler RestfulResource) Update(writer http.ResponseWriter, request *http.Request) {
	// TODO: enforce max size
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Log.Err(err).Msg("Error reading request body")
		handler.HttpRespond(writer, request, http.StatusInternalServerError, nil)
		return
	}
	contentType := GetRequestContentType(request)
	if contentType == "" {
		handler.HttpRespond(writer, request, http.StatusUnsupportedMediaType, nil)
		return
	}
	// TODO: better validation
	resource, err := handler.Resource.Unmarshal(contentType, body)
	if err != nil {
		log.Log.Err(err).Msg("Error unmarshalling request body")
		handler.HttpRespond(writer, request, http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}
	if err = handler.Resource.Validate(resource, request.Method); err != nil {
		log.Log.Err(err).Msg("Invalid body format")
		handler.HttpRespond(writer, request, http.StatusBadRequest, ErrorResponse{Message: err.Error()})
		return
	}
	resource, sErr := handler.Resource.Update(resource, mux.Vars(request))
	if sErr != nil {
		handler.HttpRespond(writer, request, sErr.StatusCode, ErrorResponse{Message: sErr.Error.Error()})
		return
	}
	handler.HttpRespond(writer, request, http.StatusOK, resource)
}

func (handler RestfulResource) HttpRespond(writer http.ResponseWriter,
	request *http.Request, code int, payload interface{}) {

	if payload != nil {
		contentType := GetResponseContentType(request)
		responseBody, err := handler.Resource.Marshal(contentType, payload)
		if err != nil {
			responseBody, _ = Marshal(contentType, ErrorResponse{Message: "Error marshalling response body"})
			code = http.StatusInternalServerError
		}
		writer.Header().Set("Content-Type", contentType)
		writer.WriteHeader(code)
		writer.Write(responseBody)
	} else {
		writer.WriteHeader(code)
	}
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
