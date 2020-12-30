// protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./vehicle.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: vehicle.proto

package proto

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type VehicleVIN struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vin string `protobuf:"bytes,1,opt,name=vin,proto3" json:"vin,omitempty"`
}

func (x *VehicleVIN) Reset() {
	*x = VehicleVIN{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vehicle_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VehicleVIN) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VehicleVIN) ProtoMessage() {}

func (x *VehicleVIN) ProtoReflect() protoreflect.Message {
	mi := &file_vehicle_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VehicleVIN.ProtoReflect.Descriptor instead.
func (*VehicleVIN) Descriptor() ([]byte, []int) {
	return file_vehicle_proto_rawDescGZIP(), []int{0}
}

func (x *VehicleVIN) GetVin() string {
	if x != nil {
		return x.Vin
	}
	return ""
}

type Vehicle struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vin           string `protobuf:"bytes,1,opt,name=vin,proto3" json:"vin,omitempty"`
	Make          string `protobuf:"bytes,2,opt,name=make,proto3" json:"make,omitempty"`
	Model         string `protobuf:"bytes,3,opt,name=model,proto3" json:"model,omitempty"`
	Year          int32  `protobuf:"varint,4,opt,name=year,proto3" json:"year,omitempty"`
	ExteriorColor string `protobuf:"bytes,5,opt,name=exterior_color,json=exteriorColor,proto3" json:"exterior_color,omitempty"`
	InteriorColor string `protobuf:"bytes,6,opt,name=interior_color,json=interiorColor,proto3" json:"interior_color,omitempty"`
}

func (x *Vehicle) Reset() {
	*x = Vehicle{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vehicle_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Vehicle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Vehicle) ProtoMessage() {}

func (x *Vehicle) ProtoReflect() protoreflect.Message {
	mi := &file_vehicle_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Vehicle.ProtoReflect.Descriptor instead.
func (*Vehicle) Descriptor() ([]byte, []int) {
	return file_vehicle_proto_rawDescGZIP(), []int{1}
}

func (x *Vehicle) GetVin() string {
	if x != nil {
		return x.Vin
	}
	return ""
}

func (x *Vehicle) GetMake() string {
	if x != nil {
		return x.Make
	}
	return ""
}

func (x *Vehicle) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

func (x *Vehicle) GetYear() int32 {
	if x != nil {
		return x.Year
	}
	return 0
}

func (x *Vehicle) GetExteriorColor() string {
	if x != nil {
		return x.ExteriorColor
	}
	return ""
}

func (x *Vehicle) GetInteriorColor() string {
	if x != nil {
		return x.InteriorColor
	}
	return ""
}

type VehicleQuery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Query string `protobuf:"bytes,1,opt,name=query,proto3" json:"query,omitempty"`
}

func (x *VehicleQuery) Reset() {
	*x = VehicleQuery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_vehicle_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VehicleQuery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VehicleQuery) ProtoMessage() {}

func (x *VehicleQuery) ProtoReflect() protoreflect.Message {
	mi := &file_vehicle_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VehicleQuery.ProtoReflect.Descriptor instead.
func (*VehicleQuery) Descriptor() ([]byte, []int) {
	return file_vehicle_proto_rawDescGZIP(), []int{2}
}

func (x *VehicleQuery) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

var File_vehicle_proto protoreflect.FileDescriptor

var file_vehicle_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x76, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x76, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1e, 0x0a, 0x0a, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65,
	0x56, 0x49, 0x4e, 0x12, 0x10, 0x0a, 0x03, 0x76, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x76, 0x69, 0x6e, 0x22, 0xa7, 0x01, 0x0a, 0x07, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x76, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x76, 0x69, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x61, 0x6b, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6d, 0x61, 0x6b, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x12, 0x0a,
	0x04, 0x79, 0x65, 0x61, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x79, 0x65, 0x61,
	0x72, 0x12, 0x25, 0x0a, 0x0e, 0x65, 0x78, 0x74, 0x65, 0x72, 0x69, 0x6f, 0x72, 0x5f, 0x63, 0x6f,
	0x6c, 0x6f, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x65, 0x78, 0x74, 0x65, 0x72,
	0x69, 0x6f, 0x72, 0x43, 0x6f, 0x6c, 0x6f, 0x72, 0x12, 0x25, 0x0a, 0x0e, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x69, 0x6f, 0x72, 0x5f, 0x63, 0x6f, 0x6c, 0x6f, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x69, 0x6f, 0x72, 0x43, 0x6f, 0x6c, 0x6f, 0x72, 0x22,
	0x24, 0x0a, 0x0c, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x32, 0xf0, 0x02, 0x0a, 0x0c, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c,
	0x65, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x35, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x56, 0x65, 0x68,
	0x69, 0x63, 0x6c, 0x65, 0x12, 0x13, 0x2e, 0x76, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x56,
	0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x56, 0x49, 0x4e, 0x1a, 0x10, 0x2e, 0x76, 0x65, 0x68, 0x69,
	0x63, 0x6c, 0x65, 0x2e, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x22, 0x00, 0x12, 0x35, 0x0a,
	0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x10,
	0x2e, 0x76, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65,
	0x1a, 0x10, 0x2e, 0x76, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x56, 0x65, 0x68, 0x69, 0x63,
	0x6c, 0x65, 0x22, 0x00, 0x12, 0x35, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x56, 0x65,
	0x68, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x10, 0x2e, 0x76, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x2e,
	0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x1a, 0x10, 0x2e, 0x76, 0x65, 0x68, 0x69, 0x63, 0x6c,
	0x65, 0x2e, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x22, 0x00, 0x12, 0x3e, 0x0a, 0x0d, 0x44,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x13, 0x2e, 0x76,
	0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x56, 0x49,
	0x4e, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x3c, 0x0a, 0x0c, 0x4c,
	0x69, 0x73, 0x74, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x1a, 0x10, 0x2e, 0x76, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x56, 0x65,
	0x68, 0x69, 0x63, 0x6c, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x3d, 0x0a, 0x0e, 0x53, 0x65, 0x61,
	0x72, 0x63, 0x68, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x12, 0x15, 0x2e, 0x76, 0x65,
	0x68, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x56, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x51, 0x75, 0x65,
	0x72, 0x79, 0x1a, 0x10, 0x2e, 0x76, 0x65, 0x68, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x56, 0x65, 0x68,
	0x69, 0x63, 0x6c, 0x65, 0x22, 0x00, 0x30, 0x01, 0x42, 0x29, 0x5a, 0x27, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x6f, 0x64, 0x65, 0x6e, 0x72, 0x2f, 0x76, 0x65,
	0x68, 0x69, 0x63, 0x6c, 0x65, 0x2d, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x76, 0x72, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_vehicle_proto_rawDescOnce sync.Once
	file_vehicle_proto_rawDescData = file_vehicle_proto_rawDesc
)

func file_vehicle_proto_rawDescGZIP() []byte {
	file_vehicle_proto_rawDescOnce.Do(func() {
		file_vehicle_proto_rawDescData = protoimpl.X.CompressGZIP(file_vehicle_proto_rawDescData)
	})
	return file_vehicle_proto_rawDescData
}

var file_vehicle_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_vehicle_proto_goTypes = []interface{}{
	(*VehicleVIN)(nil),    // 0: vehicle.VehicleVIN
	(*Vehicle)(nil),       // 1: vehicle.Vehicle
	(*VehicleQuery)(nil),  // 2: vehicle.VehicleQuery
	(*emptypb.Empty)(nil), // 3: google.protobuf.Empty
}
var file_vehicle_proto_depIdxs = []int32{
	0, // 0: vehicle.VehicleStore.GetVehicle:input_type -> vehicle.VehicleVIN
	1, // 1: vehicle.VehicleStore.CreateVehicle:input_type -> vehicle.Vehicle
	1, // 2: vehicle.VehicleStore.UpdateVehicle:input_type -> vehicle.Vehicle
	0, // 3: vehicle.VehicleStore.DeleteVehicle:input_type -> vehicle.VehicleVIN
	3, // 4: vehicle.VehicleStore.ListVehicles:input_type -> google.protobuf.Empty
	2, // 5: vehicle.VehicleStore.SearchVehicles:input_type -> vehicle.VehicleQuery
	1, // 6: vehicle.VehicleStore.GetVehicle:output_type -> vehicle.Vehicle
	1, // 7: vehicle.VehicleStore.CreateVehicle:output_type -> vehicle.Vehicle
	1, // 8: vehicle.VehicleStore.UpdateVehicle:output_type -> vehicle.Vehicle
	3, // 9: vehicle.VehicleStore.DeleteVehicle:output_type -> google.protobuf.Empty
	1, // 10: vehicle.VehicleStore.ListVehicles:output_type -> vehicle.Vehicle
	1, // 11: vehicle.VehicleStore.SearchVehicles:output_type -> vehicle.Vehicle
	6, // [6:12] is the sub-list for method output_type
	0, // [0:6] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_vehicle_proto_init() }
func file_vehicle_proto_init() {
	if File_vehicle_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_vehicle_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VehicleVIN); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_vehicle_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Vehicle); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_vehicle_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VehicleQuery); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_vehicle_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_vehicle_proto_goTypes,
		DependencyIndexes: file_vehicle_proto_depIdxs,
		MessageInfos:      file_vehicle_proto_msgTypes,
	}.Build()
	File_vehicle_proto = out.File
	file_vehicle_proto_rawDesc = nil
	file_vehicle_proto_goTypes = nil
	file_vehicle_proto_depIdxs = nil
}