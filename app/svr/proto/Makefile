.PHONY: all

all:
	protoc --gogo_out=plugins=grpc:. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf -I=. ./vehicle.proto
	protoc --gogo_out=. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf -I=. ./err.proto

# TODO: add install target
# requires gogo: https://github.com/gogo/protobuf
