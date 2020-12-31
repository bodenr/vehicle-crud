// Package main contains the entry point for the application.
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bodenr/vehicle-api/config"
	"github.com/bodenr/vehicle-api/db"
	"github.com/bodenr/vehicle-api/log"
	"github.com/bodenr/vehicle-api/resources"
	"github.com/bodenr/vehicle-api/svr"
)

func startRestApi(conf *config.HTTPConfig) <-chan bool {
	serverStop := make(chan bool, 1)
	sigStop := make(chan os.Signal)
	signal.Notify(sigStop, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

	server := svr.NewRestServer(conf, resources.StoredVehicle{})

	go func() {
		if err := server.Run(); err != nil {
			panic(err)
		}
	}()

	go server.WaitForShutdown(sigStop, serverStop)
	return serverStop
}

func startGrpcServer(conf *config.GrpcConfig) <-chan bool {
	serverStop := make(chan bool, 1)
	sigStop := make(chan os.Signal)
	signal.Notify(sigStop, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

	handler := svr.GrpcHandler{
		Resource: resources.StoredVehicle{},
	}
	server, err := svr.NewGrpcServer(conf, &handler)
	if err != nil {
		log.Log.Err(err).Msg("Failed to create grpc server")
		panic(err)
	}

	go func() {
		if err := server.Run(); err != nil {
			log.Log.Err(err).Msg("Failed to start grpc server")
			panic(err)
		}
	}()

	go server.WaitForShutdown(sigStop, serverStop)
	return serverStop
}

func main() {
	log.Log.Info().Msg("Service starting")

	// init database
	dbConfig := config.DatabaseConfig{
		DatabaseName:   "vehicles",
		Username:       "goapp",
		Port:           "5432",
		ConnectRetries: 5,
		ConnectBackoff: time.Duration(1) * time.Second,
	}
	dbConfig.Load()
	// NB: it can take up to a few seconds until the database is accepting connections when
	// started using docker compose, so hold off on starting the servers until we're sure we
	// can connect to the database
	if err := db.Initialize(&dbConfig); err != nil {
		log.Log.Err(err).Msg("Failed to initialize database")
		panic(err)
	}
	resources.StoredVehicle{}.CreateSchema()
	defer db.Close()

	// init rest api server
	httpConfig := config.HTTPConfig{
		Address: ":8080",
	}
	httpConfig.Load()
	httpStopped := startRestApi(&httpConfig)

	// init grpc server
	grpcConf := config.GrpcConfig{
		Address: ":10010",
	}
	grpcConf.Load()
	grpcStopped := startGrpcServer(&grpcConf)

	// wait for server stop
	<-httpStopped
	<-grpcStopped

	log.Log.Info().Msg("Stopped service")
}
