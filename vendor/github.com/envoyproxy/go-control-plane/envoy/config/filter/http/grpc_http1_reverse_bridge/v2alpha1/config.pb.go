// Code generated by protoc-gen-go. DO NOT EDIT.
// source: envoy/config/filter/http/grpc_http1_reverse_bridge/v2alpha1/config.proto

package envoy_config_filter_http_grpc_http1_reverse_bridge_v2alpha1

import (
	fmt "fmt"
	_ "github.com/cncf/udpa/go/udpa/annotations"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	proto "github.com/golang/protobuf/proto"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type FilterConfig struct {
	ContentType          string   `protobuf:"bytes,1,opt,name=content_type,json=contentType,proto3" json:"content_type,omitempty"`
	WithholdGrpcFrames   bool     `protobuf:"varint,2,opt,name=withhold_grpc_frames,json=withholdGrpcFrames,proto3" json:"withhold_grpc_frames,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FilterConfig) Reset()         { *m = FilterConfig{} }
func (m *FilterConfig) String() string { return proto.CompactTextString(m) }
func (*FilterConfig) ProtoMessage()    {}
func (*FilterConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a86db517160ad0a, []int{0}
}

func (m *FilterConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FilterConfig.Unmarshal(m, b)
}
func (m *FilterConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FilterConfig.Marshal(b, m, deterministic)
}
func (m *FilterConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FilterConfig.Merge(m, src)
}
func (m *FilterConfig) XXX_Size() int {
	return xxx_messageInfo_FilterConfig.Size(m)
}
func (m *FilterConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_FilterConfig.DiscardUnknown(m)
}

var xxx_messageInfo_FilterConfig proto.InternalMessageInfo

func (m *FilterConfig) GetContentType() string {
	if m != nil {
		return m.ContentType
	}
	return ""
}

func (m *FilterConfig) GetWithholdGrpcFrames() bool {
	if m != nil {
		return m.WithholdGrpcFrames
	}
	return false
}

type FilterConfigPerRoute struct {
	Disabled             bool     `protobuf:"varint,1,opt,name=disabled,proto3" json:"disabled,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FilterConfigPerRoute) Reset()         { *m = FilterConfigPerRoute{} }
func (m *FilterConfigPerRoute) String() string { return proto.CompactTextString(m) }
func (*FilterConfigPerRoute) ProtoMessage()    {}
func (*FilterConfigPerRoute) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a86db517160ad0a, []int{1}
}

func (m *FilterConfigPerRoute) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FilterConfigPerRoute.Unmarshal(m, b)
}
func (m *FilterConfigPerRoute) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FilterConfigPerRoute.Marshal(b, m, deterministic)
}
func (m *FilterConfigPerRoute) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FilterConfigPerRoute.Merge(m, src)
}
func (m *FilterConfigPerRoute) XXX_Size() int {
	return xxx_messageInfo_FilterConfigPerRoute.Size(m)
}
func (m *FilterConfigPerRoute) XXX_DiscardUnknown() {
	xxx_messageInfo_FilterConfigPerRoute.DiscardUnknown(m)
}

var xxx_messageInfo_FilterConfigPerRoute proto.InternalMessageInfo

func (m *FilterConfigPerRoute) GetDisabled() bool {
	if m != nil {
		return m.Disabled
	}
	return false
}

func init() {
	proto.RegisterType((*FilterConfig)(nil), "envoy.config.filter.http.grpc_http1_reverse_bridge.v2alpha1.FilterConfig")
	proto.RegisterType((*FilterConfigPerRoute)(nil), "envoy.config.filter.http.grpc_http1_reverse_bridge.v2alpha1.FilterConfigPerRoute")
}

func init() {
	proto.RegisterFile("envoy/config/filter/http/grpc_http1_reverse_bridge/v2alpha1/config.proto", fileDescriptor_2a86db517160ad0a)
}

var fileDescriptor_2a86db517160ad0a = []byte{
	// 339 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x91, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x86, 0x49, 0xd1, 0x5a, 0xd3, 0x1e, 0x24, 0x14, 0x2c, 0x05, 0xa5, 0xf4, 0x54, 0x3c, 0xec,
	0xda, 0xf6, 0xa6, 0x9e, 0x22, 0x54, 0xbd, 0x95, 0xe0, 0x3d, 0x6c, 0x93, 0x69, 0xb2, 0x90, 0xee,
	0x2e, 0xbb, 0xd3, 0xd8, 0xdc, 0x7c, 0x03, 0xaf, 0x3e, 0x80, 0x4f, 0xe1, 0x13, 0x78, 0xf5, 0x55,
	0x3c, 0x7a, 0x10, 0xc9, 0x26, 0x15, 0x41, 0xc4, 0x83, 0xb7, 0x49, 0x3e, 0xe6, 0x9b, 0x99, 0x7f,
	0xdd, 0x6b, 0x10, 0xb9, 0x2c, 0x68, 0x24, 0xc5, 0x92, 0x27, 0x74, 0xc9, 0x33, 0x04, 0x4d, 0x53,
	0x44, 0x45, 0x13, 0xad, 0xa2, 0xb0, 0xac, 0xc6, 0xa1, 0x86, 0x1c, 0xb4, 0x81, 0x70, 0xa1, 0x79,
	0x9c, 0x00, 0xcd, 0x27, 0x2c, 0x53, 0x29, 0x1b, 0xd7, 0x5d, 0x44, 0x69, 0x89, 0xd2, 0x3b, 0xb7,
	0x26, 0x52, 0xff, 0xab, 0x4c, 0xa4, 0xec, 0x27, 0xbf, 0x9a, 0xc8, 0xd6, 0xd4, 0x3f, 0x5e, 0xc7,
	0x8a, 0x51, 0x26, 0x84, 0x44, 0x86, 0x5c, 0x0a, 0x43, 0x57, 0x3c, 0xd1, 0x0c, 0xa1, 0x92, 0xf7,
	0x8f, 0x7e, 0x70, 0x83, 0x0c, 0xd7, 0xa6, 0xc6, 0x87, 0x39, 0xcb, 0x78, 0xcc, 0x10, 0xe8, 0xb6,
	0xa8, 0xc0, 0x30, 0x73, 0x3b, 0x33, 0xbb, 0xc9, 0xa5, 0x5d, 0xcb, 0x3b, 0x71, 0x3b, 0x91, 0x14,
	0x08, 0x02, 0x43, 0x2c, 0x14, 0xf4, 0x9c, 0x81, 0x33, 0xda, 0xf7, 0xf7, 0xde, 0xfd, 0x1d, 0xdd,
	0x18, 0x38, 0x41, 0xbb, 0x86, 0xb7, 0x85, 0x02, 0xef, 0xd4, 0xed, 0xde, 0x71, 0x4c, 0x53, 0x99,
	0xc5, 0xa1, 0x3d, 0x61, 0xa9, 0xd9, 0x0a, 0x4c, 0xaf, 0x31, 0x70, 0x46, 0xad, 0xc0, 0xdb, 0xb2,
	0x2b, 0xad, 0xa2, 0x99, 0x25, 0xc3, 0x89, 0xdb, 0xfd, 0x3e, 0x6d, 0x0e, 0x3a, 0x90, 0x6b, 0x04,
	0xaf, 0xef, 0xb6, 0x62, 0x6e, 0xd8, 0x22, 0x83, 0xd8, 0x4e, 0x6c, 0x05, 0x5f, 0xdf, 0xfe, 0x93,
	0xf3, 0xf6, 0xf8, 0xf1, 0xb0, 0x7b, 0xe1, 0x9d, 0x55, 0xf9, 0xc1, 0x06, 0x41, 0x98, 0xf2, 0xc4,
	0x3a, 0x43, 0xf3, 0x67, 0x88, 0xd3, 0xe7, 0xfb, 0x97, 0xd7, 0x66, 0xe3, 0xc0, 0x71, 0x6f, 0xb8,
	0x24, 0x56, 0xa3, 0xb4, 0xdc, 0x14, 0xe4, 0x1f, 0x2f, 0xe2, 0xb7, 0xeb, 0x1b, 0xca, 0x18, 0xe7,
	0xce, 0xa2, 0x69, 0xf3, 0x9c, 0x7e, 0x06, 0x00, 0x00, 0xff, 0xff, 0x82, 0xa5, 0x64, 0x26, 0x30,
	0x02, 0x00, 0x00,
}