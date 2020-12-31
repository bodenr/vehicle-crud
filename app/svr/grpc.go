package svr

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/bodenr/vehicle-api/config"

	"github.com/bodenr/vehicle-api/log"
	"github.com/bodenr/vehicle-api/svr/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcHandler wraps GRPC handling for the given StoredResource.
type GrpcHandler struct {
	proto.UnimplementedVehicleStoreServer
	Resource StoredResource
}

// GrpcServer the GRPC server and listener.
type GrpcServer struct {
	Server   *grpc.Server
	Listener net.Listener
}

// NewGrpcServer creates a new GrpcServer for the given config and handler.
func NewGrpcServer(conf *config.GrpcConfig, handler *GrpcHandler) (*GrpcServer, error) {
	listener, err := net.Listen("tcp", conf.Address)
	if err != nil {
		return nil, err
	}
	log.Log.Info().Str(log.Hostname, conf.Address).Msg("created grpc listener")

	// TODO: TLS and all that other goodness
	server := grpc.NewServer()
	proto.RegisterVehicleStoreServer(server, handler)

	return &GrpcServer{
		Server:   server,
		Listener: listener,
	}, nil
}

// Run starts the GRPC server and is blocking.
func (server *GrpcServer) Run() error {
	log.Log.Info().Msg("starting grpc server")
	return server.Server.Serve(server.Listener)
}

// WaitForShutdown waits on the terminate channel and shuts down the GRPC server upon signal.
func (server *GrpcServer) WaitForShutdown(terminate <-chan os.Signal, done chan bool) {
	termSig := <-terminate

	log.Log.Info().Str(log.Signal, termSig.String()).Msg("shutting down grpc server")
	server.Server.GracefulStop()
	log.Log.Info().Msg("grpc server gracefully stopped")
	close(done)
}

// GetVehicle handle getting a vehicle for GRPC.
func (handler *GrpcHandler) GetVehicle(ctx context.Context, vin *proto.VehicleVIN) (*proto.Vehicle, error) {
	vars := map[string]string{
		"vin": vin.GetVin(),
	}
	resource, err := handler.Resource.Get(vars)
	if err != nil {
		log.Log.Err(err.Error).Msg("Error getting vehicle")
		code := codes.Internal
		if err.StatusCode == http.StatusNotFound {
			code = codes.NotFound
		}
		return nil, status.Error(code, err.Error.Error())
	}

	v := resource.(proto.Vehicle)
	return &v, nil
}

// CreateVehicle handler creating a vehicle over GRPC.
func (handler *GrpcHandler) CreateVehicle(ctx context.Context, vehicle *proto.Vehicle) (*proto.Vehicle, error) {
	if err := handler.Resource.Validate(vehicle, http.MethodPost); err != nil {
		log.Log.Err(err).Msg("Invalid format")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	storedResource, sErr := handler.Resource.Create(vehicle)
	if sErr != nil {
		log.Log.Err(sErr.Error).Msg("Error creating vehicle")
		code := codes.Internal
		if sErr.StatusCode == http.StatusBadRequest {
			code = codes.AlreadyExists
		}
		return nil, status.Error(code, sErr.Error.Error())
	}
	v := storedResource.(proto.Vehicle)
	return &v, nil
}

// UpdateVehicle handle updating a vehicle over GRPC.
func (handler *GrpcHandler) UpdateVehicle(ctx context.Context, vehicle *proto.Vehicle) (*proto.Vehicle, error) {

	if err := handler.Resource.Validate(vehicle, http.MethodPut); err != nil {
		log.Log.Err(err).Msg("Invalid vehicle format")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	vars := map[string]string{
		"vin": vehicle.Vin,
	}
	storedResource, sErr := handler.Resource.Update(vehicle, vars)
	if sErr != nil {
		log.Log.Err(sErr.Error).Msg("Error updating vehicle")
		return nil, status.Error(codes.Internal, sErr.Error.Error())
	}
	v := storedResource.(proto.Vehicle)
	return &v, nil
}

// DeleteVehicle handles deleting a vehicle over GRPC.
func (handler *GrpcHandler) DeleteVehicle(ctx context.Context, vehicleVin *proto.VehicleVIN) (*proto.EmptyMessage, error) {
	vars := map[string]string{
		"vin": vehicleVin.Vin,
	}

	if err := handler.Resource.Delete(vars); err != nil {
		log.Log.Err(err.Error).Str(log.VIN, vehicleVin.Vin).Msg("Error deleting vehicle")
		return nil, status.Error(codes.Internal, err.Error.Error())
	}
	return &proto.EmptyMessage{}, nil
}

// ListVehicles handles listing vehicles over GRPC.
func (handler *GrpcHandler) ListVehicles(e *proto.EmptyMessage, stream proto.VehicleStore_ListVehiclesServer) error {
	resources, sErr := handler.Resource.List()
	if sErr != nil {
		return status.Error(codes.Internal, sErr.Error.Error())
	}
	for _, resource := range resources {
		v := resource.(proto.Vehicle)
		if err := stream.Send(&v); err != nil {
			return err
		}
	}
	return nil
}

// SearchVehicles handles searching for vehciles over GRPC.
func (handler *GrpcHandler) SearchVehicles(query *proto.VehicleQuery, stream proto.VehicleStore_SearchVehiclesServer) error {
	queryValues, err := url.ParseQuery(query.Query)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	resources, sErr := handler.Resource.Search(queryValues)
	if sErr != nil {
		code := codes.Internal
		if sErr.StatusCode == http.StatusBadRequest {
			code = codes.InvalidArgument
		}
		return status.Error(code, sErr.Error.Error())
	}
	for _, resource := range resources {
		v := resource.(proto.Vehicle)
		if err = stream.Send(&v); err != nil {
			return err
		}
	}
	return nil
}
