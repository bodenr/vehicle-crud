package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bodenr/vehicle-app/config"
	"github.com/bodenr/vehicle-app/db"
	"github.com/bodenr/vehicle-app/log"
	"github.com/bodenr/vehicle-app/resources"
	"github.com/bodenr/vehicle-app/svr"
)

func startRestApi(conf *config.HTTPConfig) <-chan bool {
	serverStop := make(chan bool, 1)
	sigStop := make(chan os.Signal)
	signal.Notify(sigStop, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

	server := svr.NewRestServer(conf, resources.BindVehicleRequestHandlers)
	go server.WaitForShutdown(sigStop, serverStop)

	log.Log.Info().Msg("starting http server on port " + server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Log.Err(err).Msg("failed to start http server")
		panic(err)
	}

	return serverStop
}

func main() {
	log.Log.Info().Msg("service starting")

	// init database
	dbConfig := config.DatabaseConfig{
		DatabaseName:   "vehicles",
		Username:       "goapp",
		Port:           "5432",
		ConnectRetries: 5,
		ConnectBackoff: time.Duration(1) * time.Second,
	}
	dbConfig.Load()
	if err := db.Initialize(&dbConfig); err != nil {
		log.Log.Err(err).Msg("failed to initialize database")
		panic(err)
	}
	if err := db.GetDB().AutoMigrate(&resources.VehicleModel{}); err != nil {
		log.Log.Err(err).Msg("database schema migration failed")
		panic(err)
	}
	defer db.Close()

	// init rest api server
	httpConfig := config.HTTPConfig{
		Address: ":8080",
	}
	httpConfig.Load()
	httpStopped := startRestApi(&httpConfig)

	// wait for server stop
	<-httpStopped

	log.Log.Info().Msg("stopping service")
}
