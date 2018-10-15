// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ntx/hello/hello.proto

/*
	Package hello is a generated protocol buffer package.

	It is generated from these files:
		ntx/hello/hello.proto

	It has these top-level messages:
		Hello
*/
package hello

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Hello struct {
	Msg   string `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	Times int32  `protobuf:"varint,2,opt,name=times,proto3" json:"times,omitempty"`
}

func (m *Hello) Reset()                    { *m = Hello{} }
func (m *Hello) String() string            { return proto.CompactTextString(m) }
func (*Hello) ProtoMessage()               {}
func (*Hello) Descriptor() ([]byte, []int) { return fileDescriptorHello, []int{0} }

func (m *Hello) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func (m *Hello) GetTimes() int32 {
	if m != nil {
		return m.Times
	}
	return 0
}

func init() {
	proto.RegisterType((*Hello)(nil), "ntx.hello.Hello")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for HelloService service

type HelloServiceClient interface {
	SayHello(ctx context.Context, in *Hello, opts ...grpc.CallOption) (*Hello, error)
	SayHelloNTimes(ctx context.Context, in *Hello, opts ...grpc.CallOption) (HelloService_SayHelloNTimesClient, error)
	SayHelloBidi(ctx context.Context, opts ...grpc.CallOption) (HelloService_SayHelloBidiClient, error)
}

type helloServiceClient struct {
	cc *grpc.ClientConn
}

func NewHelloServiceClient(cc *grpc.ClientConn) HelloServiceClient {
	return &helloServiceClient{cc}
}

func (c *helloServiceClient) SayHello(ctx context.Context, in *Hello, opts ...grpc.CallOption) (*Hello, error) {
	out := new(Hello)
	err := grpc.Invoke(ctx, "/ntx.hello.HelloService/SayHello", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helloServiceClient) SayHelloNTimes(ctx context.Context, in *Hello, opts ...grpc.CallOption) (HelloService_SayHelloNTimesClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_HelloService_serviceDesc.Streams[0], c.cc, "/ntx.hello.HelloService/SayHelloNTimes", opts...)
	if err != nil {
		return nil, err
	}
	x := &helloServiceSayHelloNTimesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type HelloService_SayHelloNTimesClient interface {
	Recv() (*Hello, error)
	grpc.ClientStream
}

type helloServiceSayHelloNTimesClient struct {
	grpc.ClientStream
}

func (x *helloServiceSayHelloNTimesClient) Recv() (*Hello, error) {
	m := new(Hello)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *helloServiceClient) SayHelloBidi(ctx context.Context, opts ...grpc.CallOption) (HelloService_SayHelloBidiClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_HelloService_serviceDesc.Streams[1], c.cc, "/ntx.hello.HelloService/SayHelloBidi", opts...)
	if err != nil {
		return nil, err
	}
	x := &helloServiceSayHelloBidiClient{stream}
	return x, nil
}

type HelloService_SayHelloBidiClient interface {
	Send(*Hello) error
	Recv() (*Hello, error)
	grpc.ClientStream
}

type helloServiceSayHelloBidiClient struct {
	grpc.ClientStream
}

func (x *helloServiceSayHelloBidiClient) Send(m *Hello) error {
	return x.ClientStream.SendMsg(m)
}

func (x *helloServiceSayHelloBidiClient) Recv() (*Hello, error) {
	m := new(Hello)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for HelloService service

type HelloServiceServer interface {
	SayHello(context.Context, *Hello) (*Hello, error)
	SayHelloNTimes(*Hello, HelloService_SayHelloNTimesServer) error
	SayHelloBidi(HelloService_SayHelloBidiServer) error
}

func RegisterHelloServiceServer(s *grpc.Server, srv HelloServiceServer) {
	s.RegisterService(&_HelloService_serviceDesc, srv)
}

func _HelloService_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Hello)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServiceServer).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ntx.hello.HelloService/SayHello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelloServiceServer).SayHello(ctx, req.(*Hello))
	}
	return interceptor(ctx, in, info, handler)
}

func _HelloService_SayHelloNTimes_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Hello)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HelloServiceServer).SayHelloNTimes(m, &helloServiceSayHelloNTimesServer{stream})
}

type HelloService_SayHelloNTimesServer interface {
	Send(*Hello) error
	grpc.ServerStream
}

type helloServiceSayHelloNTimesServer struct {
	grpc.ServerStream
}

func (x *helloServiceSayHelloNTimesServer) Send(m *Hello) error {
	return x.ServerStream.SendMsg(m)
}

func _HelloService_SayHelloBidi_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(HelloServiceServer).SayHelloBidi(&helloServiceSayHelloBidiServer{stream})
}

type HelloService_SayHelloBidiServer interface {
	Send(*Hello) error
	Recv() (*Hello, error)
	grpc.ServerStream
}

type helloServiceSayHelloBidiServer struct {
	grpc.ServerStream
}

func (x *helloServiceSayHelloBidiServer) Send(m *Hello) error {
	return x.ServerStream.SendMsg(m)
}

func (x *helloServiceSayHelloBidiServer) Recv() (*Hello, error) {
	m := new(Hello)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _HelloService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ntx.hello.HelloService",
	HandlerType: (*HelloServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _HelloService_SayHello_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SayHelloNTimes",
			Handler:       _HelloService_SayHelloNTimes_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SayHelloBidi",
			Handler:       _HelloService_SayHelloBidi_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "ntx/hello/hello.proto",
}

func (m *Hello) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Hello) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Msg) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintHello(dAtA, i, uint64(len(m.Msg)))
		i += copy(dAtA[i:], m.Msg)
	}
	if m.Times != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintHello(dAtA, i, uint64(m.Times))
	}
	return i, nil
}

func encodeFixed64Hello(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Hello(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintHello(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Hello) Size() (n int) {
	var l int
	_ = l
	l = len(m.Msg)
	if l > 0 {
		n += 1 + l + sovHello(uint64(l))
	}
	if m.Times != 0 {
		n += 1 + sovHello(uint64(m.Times))
	}
	return n
}

func sovHello(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozHello(x uint64) (n int) {
	return sovHello(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Hello) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowHello
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Hello: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Hello: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Msg", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHello
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthHello
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Msg = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Times", wireType)
			}
			m.Times = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowHello
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Times |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipHello(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthHello
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipHello(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowHello
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowHello
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowHello
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthHello
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowHello
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipHello(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthHello = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowHello   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("ntx/hello/hello.proto", fileDescriptorHello) }

var fileDescriptorHello = []byte{
	// 196 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xcd, 0x2b, 0xa9, 0xd0,
	0xcf, 0x48, 0xcd, 0xc9, 0xc9, 0x87, 0x90, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42, 0x9c, 0x79,
	0x25, 0x15, 0x7a, 0x60, 0x01, 0x25, 0x7d, 0x2e, 0x56, 0x0f, 0x10, 0x43, 0x48, 0x80, 0x8b, 0x39,
	0xb7, 0x38, 0x5d, 0x82, 0x51, 0x81, 0x51, 0x83, 0x33, 0x08, 0xc4, 0x14, 0x12, 0xe1, 0x62, 0x2d,
	0xc9, 0xcc, 0x4d, 0x2d, 0x96, 0x60, 0x52, 0x60, 0xd4, 0x60, 0x0d, 0x82, 0x70, 0x8c, 0xd6, 0x31,
	0x72, 0xf1, 0x80, 0x75, 0x04, 0xa7, 0x16, 0x95, 0x65, 0x26, 0xa7, 0x0a, 0xe9, 0x71, 0x71, 0x04,
	0x27, 0x56, 0x42, 0x0d, 0xd1, 0x83, 0x9b, 0xac, 0x07, 0x16, 0x91, 0xc2, 0x10, 0x11, 0x32, 0xe3,
	0xe2, 0x83, 0xa9, 0xf7, 0x0b, 0x01, 0x19, 0x49, 0x8c, 0x2e, 0x03, 0x46, 0x21, 0x33, 0x2e, 0x1e,
	0x98, 0x3e, 0xa7, 0xcc, 0x94, 0x4c, 0x62, 0x74, 0x69, 0x30, 0x1a, 0x30, 0x3a, 0x39, 0x9c, 0x78,
	0x24, 0xc7, 0x78, 0xe1, 0x91, 0x1c, 0xe3, 0x83, 0x47, 0x72, 0x8c, 0x33, 0x1e, 0xcb, 0x31, 0x70,
	0x09, 0x25, 0x57, 0x81, 0x55, 0x82, 0x03, 0x03, 0xa2, 0x3e, 0x8a, 0x15, 0x4c, 0xad, 0x62, 0xc2,
	0x22, 0x97, 0xc4, 0x06, 0xe6, 0x18, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0xe0, 0x46, 0x7a, 0xf4,
	0x4e, 0x01, 0x00, 0x00,
}