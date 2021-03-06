// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: err.proto

package proto

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type ErrorResponse struct {
	Message              string   `protobuf:"bytes,1,opt,name=Message,proto3" json:"error_message" xml:"error_message"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ErrorResponse) Reset()         { *m = ErrorResponse{} }
func (m *ErrorResponse) String() string { return proto.CompactTextString(m) }
func (*ErrorResponse) ProtoMessage()    {}
func (*ErrorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_b4a1db73bc95ee8c, []int{0}
}
func (m *ErrorResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ErrorResponse.Unmarshal(m, b)
}
func (m *ErrorResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ErrorResponse.Marshal(b, m, deterministic)
}
func (m *ErrorResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ErrorResponse.Merge(m, src)
}
func (m *ErrorResponse) XXX_Size() int {
	return xxx_messageInfo_ErrorResponse.Size(m)
}
func (m *ErrorResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ErrorResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ErrorResponse proto.InternalMessageInfo

func (m *ErrorResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*ErrorResponse)(nil), "err.ErrorResponse")
}

func init() { proto.RegisterFile("err.proto", fileDescriptor_b4a1db73bc95ee8c) }

var fileDescriptor_b4a1db73bc95ee8c = []byte{
	// 147 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4c, 0x2d, 0x2a, 0xd2,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4e, 0x2d, 0x2a, 0x92, 0xd2, 0x4d, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0xcf, 0x4f, 0xcf, 0xd7, 0x07, 0xcb, 0x25, 0x95,
	0xa6, 0x81, 0x79, 0x60, 0x0e, 0x98, 0x05, 0xd1, 0xa3, 0x14, 0xcc, 0xc5, 0xeb, 0x5a, 0x54, 0x94,
	0x5f, 0x14, 0x94, 0x5a, 0x5c, 0x90, 0x9f, 0x57, 0x9c, 0x2a, 0xe4, 0xc4, 0xc5, 0xee, 0x9b, 0x5a,
	0x5c, 0x9c, 0x98, 0x9e, 0x2a, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0xe9, 0xa4, 0xf1, 0xea, 0x9e, 0x3c,
	0x6f, 0x2a, 0x48, 0x4d, 0x7c, 0x2e, 0x44, 0xe2, 0xd3, 0x3d, 0x79, 0xe1, 0x8a, 0xdc, 0x1c, 0x2b,
	0x25, 0x14, 0x51, 0xa5, 0x20, 0x98, 0x46, 0x27, 0xee, 0x1f, 0x0f, 0xe5, 0x18, 0xa3, 0x58, 0x21,
	0x36, 0xb3, 0x81, 0x29, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x29, 0x5f, 0x90, 0x05, 0xa9,
	0x00, 0x00, 0x00,
}

func NewPopulatedErrorResponse(r randyErr, easy bool) *ErrorResponse {
	this := &ErrorResponse{}
	this.Message = string(randStringErr(r))
	if !easy && r.Intn(10) != 0 {
		this.XXX_unrecognized = randUnrecognizedErr(r, 2)
	}
	return this
}

type randyErr interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneErr(r randyErr) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringErr(r randyErr) string {
	v1 := r.Intn(100)
	tmps := make([]rune, v1)
	for i := 0; i < v1; i++ {
		tmps[i] = randUTF8RuneErr(r)
	}
	return string(tmps)
}
func randUnrecognizedErr(r randyErr, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldErr(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldErr(dAtA []byte, r randyErr, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateErr(dAtA, uint64(key))
		v2 := r.Int63()
		if r.Intn(2) == 0 {
			v2 *= -1
		}
		dAtA = encodeVarintPopulateErr(dAtA, uint64(v2))
	case 1:
		dAtA = encodeVarintPopulateErr(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateErr(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateErr(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateErr(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateErr(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
