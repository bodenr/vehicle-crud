package svr

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/bodenr/vehicle-app/config"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/hlog"

	"github.com/bodenr/vehicle-app/log"
)

type Error struct {
	Message string
}

type RestServer struct {
	*http.Server
}

type BindRequestHandler func(router *mux.Router)

func NewRestServer(conf *config.HTTPConfig, bindFuncs ...BindRequestHandler) *RestServer {

	router := mux.NewRouter()

	subrouter := router.PathPrefix("/api/v1").Subrouter()

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

	for _, bindFunc := range bindFuncs {
		bindFunc(subrouter)
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
