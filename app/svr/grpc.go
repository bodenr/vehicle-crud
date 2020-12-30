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
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type GrpcHandler struct {
	proto.UnimplementedVehicleStoreServer
	Resource StoredResource
}

type GrpcServer struct {
	Server   *grpc.Server
	Listener net.Listener
}

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

func (server *GrpcServer) Run() error {
	log.Log.Info().Msg("starting grpc server")
	return server.Server.Serve(server.Listener)
}

func (server *GrpcServer) WaitForShutdown(terminate <-chan os.Signal, done chan bool) {
	termSig := <-terminate

	log.Log.Info().Str(log.Signal, termSig.String()).Msg("shutting down grpc server")
	server.Server.GracefulStop()
	log.Log.Info().Msg("grpc server gracefully stopped")
	close(done)
}

// TODO: consider using gogo extensions for a single shared struct type
func resourceToProto(resource *StoredResource) (*proto.Vehicle, error) {
	bytes, err := Marshal("application/json", resource)
	if err != nil {
		return nil, err
	}
	var vehicle *proto.Vehicle
	if err = Unmarshal("application/json", bytes, vehicle); err != nil {
		return vehicle, err
	}

	return vehicle, nil
}

func protoToResource(vehicle *proto.Vehicle) (StoredResource, error) {
	bytes, err := Marshal("application/json", vehicle)
	if err != nil {
		return nil, err
	}
	var resource StoredResource
	if err = Unmarshal("application/json", bytes, resource); err != nil {
		return nil, err
	}
	return resource, nil
}

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

	v, rErr := resourceToProto(&resource)
	if rErr != nil {
		return nil, status.Error(codes.Internal, rErr.Error())
	}
	return v, nil
}

func (handler *GrpcHandler) CreateVehicle(ctx context.Context, vehicle *proto.Vehicle) (*proto.Vehicle, error) {
	resource, err := protoToResource(vehicle)
	if err != nil {
		log.Log.Err(err).Msg("Error converting protobuf to resource")
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := resource.Validate("POST"); err != nil {
		log.Log.Err(err).Msg("Invalid format")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if sErr := resource.Create(); sErr != nil {
		log.Log.Err(sErr.Error).Msg("Error creating vehicle")
		code := codes.Internal
		if sErr.StatusCode == http.StatusBadRequest {
			code = codes.AlreadyExists
		}
		return nil, status.Error(code, sErr.Error.Error())
	}
	v, err := resourceToProto(&resource)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return v, nil
}

func (handler *GrpcHandler) UpdateVehicle(ctx context.Context, vehicle *proto.Vehicle) (*proto.Vehicle, error) {
	resource, err := protoToResource(vehicle)
	if err != nil {
		log.Log.Err(err).Msg("Error converting protobuf to resource")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = resource.Validate("PUT"); err != nil {
		log.Log.Err(err).Msg("Invalid vehicle format")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	vars := map[string]string{
		"vin": vehicle.Vin,
	}
	if sErr := resource.Update(vars); err != nil {
		log.Log.Err(sErr.Error).Msg("Error updating vehicle")
		return nil, status.Error(codes.Internal, sErr.Error.Error())
	}
	v, err := resourceToProto(&resource)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return v, nil
}

func (handler *GrpcHandler) DeleteVehicle(ctx context.Context, vehicleVin *proto.VehicleVIN) (*emptypb.Empty, error) {
	vars := map[string]string{
		"vin": vehicleVin.Vin,
	}

	if err := handler.Resource.Delete(vars); err != nil {
		log.Log.Err(err.Error).Str(log.VIN, vehicleVin.Vin).Msg("Error deleting vehicle")
		return nil, status.Error(codes.Internal, err.Error.Error())
	}
	return nil, nil
}

func (handler *GrpcHandler) ListVehicles(e *emptypb.Empty, stream proto.VehicleStore_ListVehiclesServer) error {
	resources, sErr := handler.Resource.List()
	if sErr != nil {
		return status.Error(codes.Internal, sErr.Error.Error())
	}
	for _, resource := range resources {
		v, err := resourceToProto(&resource)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		if err = stream.Send(v); err != nil {
			return err
		}
	}
	return nil
}

func (handler *GrpcHandler) SearchVehicles(query *proto.VehicleQuery, stream proto.VehicleStore_SearchVehiclesServer) error {
	queryValues, err := url.ParseQuery(query.Query)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	resources, sErr := handler.Resource.Search(&queryValues)
	if sErr != nil {
		code := codes.Internal
		if sErr.StatusCode == http.StatusBadRequest {
			code = codes.InvalidArgument
		}
		return status.Error(code, sErr.Error.Error())
	}
	for _, resource := range resources {
		v, err := resourceToProto(&resource)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		if err = stream.Send(v); err != nil {
			return err
		}
	}
	return nil
}
