// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ntx/v2t/engine/g2p.proto

package engine

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

type G2PTranscription struct {
	Grapheme string   `protobuf:"bytes,1,opt,name=grapheme,proto3" json:"grapheme,omitempty"`
	Phoneme  []string `protobuf:"bytes,2,rep,name=phoneme" json:"phoneme,omitempty"`
}

func (m *G2PTranscription) Reset()                    { *m = G2PTranscription{} }
func (m *G2PTranscription) String() string            { return proto.CompactTextString(m) }
func (*G2PTranscription) ProtoMessage()               {}
func (*G2PTranscription) Descriptor() ([]byte, []int) { return fileDescriptorG2P, []int{0} }

func (m *G2PTranscription) GetGrapheme() string {
	if m != nil {
		return m.Grapheme
	}
	return ""
}

func (m *G2PTranscription) GetPhoneme() []string {
	if m != nil {
		return m.Phoneme
	}
	return nil
}

type G2PRequest struct {
	Lang string `protobuf:"bytes,1,opt,name=lang,proto3" json:"lang,omitempty"`
	Text string `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
}

func (m *G2PRequest) Reset()                    { *m = G2PRequest{} }
func (m *G2PRequest) String() string            { return proto.CompactTextString(m) }
func (*G2PRequest) ProtoMessage()               {}
func (*G2PRequest) Descriptor() ([]byte, []int) { return fileDescriptorG2P, []int{1} }

func (m *G2PRequest) GetLang() string {
	if m != nil {
		return m.Lang
	}
	return ""
}

func (m *G2PRequest) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

type G2PResponse struct {
	Response []*G2PTranscription `protobuf:"bytes,1,rep,name=response" json:"response,omitempty"`
}

func (m *G2PResponse) Reset()                    { *m = G2PResponse{} }
func (m *G2PResponse) String() string            { return proto.CompactTextString(m) }
func (*G2PResponse) ProtoMessage()               {}
func (*G2PResponse) Descriptor() ([]byte, []int) { return fileDescriptorG2P, []int{2} }

func (m *G2PResponse) GetResponse() []*G2PTranscription {
	if m != nil {
		return m.Response
	}
	return nil
}

type G2PLanguagesRequest struct {
}

func (m *G2PLanguagesRequest) Reset()                    { *m = G2PLanguagesRequest{} }
func (m *G2PLanguagesRequest) String() string            { return proto.CompactTextString(m) }
func (*G2PLanguagesRequest) ProtoMessage()               {}
func (*G2PLanguagesRequest) Descriptor() ([]byte, []int) { return fileDescriptorG2P, []int{3} }

type G2PLanguagesResponse struct {
	Languages []string                         `protobuf:"bytes,1,rep,name=languages" json:"languages,omitempty"`
	Alphabets []*G2PLanguagesResponse_Alphabet `protobuf:"bytes,2,rep,name=alphabets" json:"alphabets,omitempty"`
}

func (m *G2PLanguagesResponse) Reset()                    { *m = G2PLanguagesResponse{} }
func (m *G2PLanguagesResponse) String() string            { return proto.CompactTextString(m) }
func (*G2PLanguagesResponse) ProtoMessage()               {}
func (*G2PLanguagesResponse) Descriptor() ([]byte, []int) { return fileDescriptorG2P, []int{4} }

func (m *G2PLanguagesResponse) GetLanguages() []string {
	if m != nil {
		return m.Languages
	}
	return nil
}

func (m *G2PLanguagesResponse) GetAlphabets() []*G2PLanguagesResponse_Alphabet {
	if m != nil {
		return m.Alphabets
	}
	return nil
}

type G2PLanguagesResponse_Alphabet struct {
	Lang         string   `protobuf:"bytes,1,opt,name=lang,proto3" json:"lang,omitempty"`
	HelpLink     string   `protobuf:"bytes,2,opt,name=helpLink,proto3" json:"helpLink,omitempty"`
	AllowedChars []string `protobuf:"bytes,3,rep,name=allowedChars" json:"allowedChars,omitempty"`
}

func (m *G2PLanguagesResponse_Alphabet) Reset()         { *m = G2PLanguagesResponse_Alphabet{} }
func (m *G2PLanguagesResponse_Alphabet) String() string { return proto.CompactTextString(m) }
func (*G2PLanguagesResponse_Alphabet) ProtoMessage()    {}
func (*G2PLanguagesResponse_Alphabet) Descriptor() ([]byte, []int) {
	return fileDescriptorG2P, []int{4, 0}
}

func (m *G2PLanguagesResponse_Alphabet) GetLang() string {
	if m != nil {
		return m.Lang
	}
	return ""
}

func (m *G2PLanguagesResponse_Alphabet) GetHelpLink() string {
	if m != nil {
		return m.HelpLink
	}
	return ""
}

func (m *G2PLanguagesResponse_Alphabet) GetAllowedChars() []string {
	if m != nil {
		return m.AllowedChars
	}
	return nil
}

func init() {
	proto.RegisterType((*G2PTranscription)(nil), "ntx.v2t.engine.G2PTranscription")
	proto.RegisterType((*G2PRequest)(nil), "ntx.v2t.engine.G2PRequest")
	proto.RegisterType((*G2PResponse)(nil), "ntx.v2t.engine.G2PResponse")
	proto.RegisterType((*G2PLanguagesRequest)(nil), "ntx.v2t.engine.G2PLanguagesRequest")
	proto.RegisterType((*G2PLanguagesResponse)(nil), "ntx.v2t.engine.G2PLanguagesResponse")
	proto.RegisterType((*G2PLanguagesResponse_Alphabet)(nil), "ntx.v2t.engine.G2PLanguagesResponse.Alphabet")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for G2PService service

type G2PServiceClient interface {
	Transcribe(ctx context.Context, in *G2PRequest, opts ...grpc.CallOption) (*G2PResponse, error)
	Languages(ctx context.Context, in *G2PLanguagesRequest, opts ...grpc.CallOption) (*G2PLanguagesResponse, error)
}

type g2PServiceClient struct {
	cc *grpc.ClientConn
}

func NewG2PServiceClient(cc *grpc.ClientConn) G2PServiceClient {
	return &g2PServiceClient{cc}
}

func (c *g2PServiceClient) Transcribe(ctx context.Context, in *G2PRequest, opts ...grpc.CallOption) (*G2PResponse, error) {
	out := new(G2PResponse)
	err := grpc.Invoke(ctx, "/ntx.v2t.engine.G2PService/Transcribe", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *g2PServiceClient) Languages(ctx context.Context, in *G2PLanguagesRequest, opts ...grpc.CallOption) (*G2PLanguagesResponse, error) {
	out := new(G2PLanguagesResponse)
	err := grpc.Invoke(ctx, "/ntx.v2t.engine.G2PService/Languages", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for G2PService service

type G2PServiceServer interface {
	Transcribe(context.Context, *G2PRequest) (*G2PResponse, error)
	Languages(context.Context, *G2PLanguagesRequest) (*G2PLanguagesResponse, error)
}

func RegisterG2PServiceServer(s *grpc.Server, srv G2PServiceServer) {
	s.RegisterService(&_G2PService_serviceDesc, srv)
}

func _G2PService_Transcribe_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(G2PRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(G2PServiceServer).Transcribe(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ntx.v2t.engine.G2PService/Transcribe",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(G2PServiceServer).Transcribe(ctx, req.(*G2PRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _G2PService_Languages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(G2PLanguagesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(G2PServiceServer).Languages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ntx.v2t.engine.G2PService/Languages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(G2PServiceServer).Languages(ctx, req.(*G2PLanguagesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _G2PService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ntx.v2t.engine.G2PService",
	HandlerType: (*G2PServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Transcribe",
			Handler:    _G2PService_Transcribe_Handler,
		},
		{
			MethodName: "Languages",
			Handler:    _G2PService_Languages_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ntx/v2t/engine/g2p.proto",
}

func (m *G2PTranscription) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *G2PTranscription) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Grapheme) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintG2P(dAtA, i, uint64(len(m.Grapheme)))
		i += copy(dAtA[i:], m.Grapheme)
	}
	if len(m.Phoneme) > 0 {
		for _, s := range m.Phoneme {
			dAtA[i] = 0x12
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	return i, nil
}

func (m *G2PRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *G2PRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Lang) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintG2P(dAtA, i, uint64(len(m.Lang)))
		i += copy(dAtA[i:], m.Lang)
	}
	if len(m.Text) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintG2P(dAtA, i, uint64(len(m.Text)))
		i += copy(dAtA[i:], m.Text)
	}
	return i, nil
}

func (m *G2PResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *G2PResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Response) > 0 {
		for _, msg := range m.Response {
			dAtA[i] = 0xa
			i++
			i = encodeVarintG2P(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *G2PLanguagesRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *G2PLanguagesRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *G2PLanguagesResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *G2PLanguagesResponse) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Languages) > 0 {
		for _, s := range m.Languages {
			dAtA[i] = 0xa
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	if len(m.Alphabets) > 0 {
		for _, msg := range m.Alphabets {
			dAtA[i] = 0x12
			i++
			i = encodeVarintG2P(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *G2PLanguagesResponse_Alphabet) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *G2PLanguagesResponse_Alphabet) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Lang) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintG2P(dAtA, i, uint64(len(m.Lang)))
		i += copy(dAtA[i:], m.Lang)
	}
	if len(m.HelpLink) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintG2P(dAtA, i, uint64(len(m.HelpLink)))
		i += copy(dAtA[i:], m.HelpLink)
	}
	if len(m.AllowedChars) > 0 {
		for _, s := range m.AllowedChars {
			dAtA[i] = 0x1a
			i++
			l = len(s)
			for l >= 1<<7 {
				dAtA[i] = uint8(uint64(l)&0x7f | 0x80)
				l >>= 7
				i++
			}
			dAtA[i] = uint8(l)
			i++
			i += copy(dAtA[i:], s)
		}
	}
	return i, nil
}

func encodeFixed64G2P(dAtA []byte, offset int, v uint64) int {
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
func encodeFixed32G2P(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintG2P(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *G2PTranscription) Size() (n int) {
	var l int
	_ = l
	l = len(m.Grapheme)
	if l > 0 {
		n += 1 + l + sovG2P(uint64(l))
	}
	if len(m.Phoneme) > 0 {
		for _, s := range m.Phoneme {
			l = len(s)
			n += 1 + l + sovG2P(uint64(l))
		}
	}
	return n
}

func (m *G2PRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.Lang)
	if l > 0 {
		n += 1 + l + sovG2P(uint64(l))
	}
	l = len(m.Text)
	if l > 0 {
		n += 1 + l + sovG2P(uint64(l))
	}
	return n
}

func (m *G2PResponse) Size() (n int) {
	var l int
	_ = l
	if len(m.Response) > 0 {
		for _, e := range m.Response {
			l = e.Size()
			n += 1 + l + sovG2P(uint64(l))
		}
	}
	return n
}

func (m *G2PLanguagesRequest) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *G2PLanguagesResponse) Size() (n int) {
	var l int
	_ = l
	if len(m.Languages) > 0 {
		for _, s := range m.Languages {
			l = len(s)
			n += 1 + l + sovG2P(uint64(l))
		}
	}
	if len(m.Alphabets) > 0 {
		for _, e := range m.Alphabets {
			l = e.Size()
			n += 1 + l + sovG2P(uint64(l))
		}
	}
	return n
}

func (m *G2PLanguagesResponse_Alphabet) Size() (n int) {
	var l int
	_ = l
	l = len(m.Lang)
	if l > 0 {
		n += 1 + l + sovG2P(uint64(l))
	}
	l = len(m.HelpLink)
	if l > 0 {
		n += 1 + l + sovG2P(uint64(l))
	}
	if len(m.AllowedChars) > 0 {
		for _, s := range m.AllowedChars {
			l = len(s)
			n += 1 + l + sovG2P(uint64(l))
		}
	}
	return n
}

func sovG2P(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozG2P(x uint64) (n int) {
	return sovG2P(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *G2PTranscription) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowG2P
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
			return fmt.Errorf("proto: G2PTranscription: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: G2PTranscription: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Grapheme", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowG2P
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
				return ErrInvalidLengthG2P
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Grapheme = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Phoneme", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowG2P
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
				return ErrInvalidLengthG2P
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Phoneme = append(m.Phoneme, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipG2P(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthG2P
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
func (m *G2PRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowG2P
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
			return fmt.Errorf("proto: G2PRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: G2PRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Lang", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowG2P
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
				return ErrInvalidLengthG2P
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Lang = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Text", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowG2P
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
				return ErrInvalidLengthG2P
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Text = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipG2P(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthG2P
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
func (m *G2PResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowG2P
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
			return fmt.Errorf("proto: G2PResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: G2PResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Response", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowG2P
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthG2P
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Response = append(m.Response, &G2PTranscription{})
			if err := m.Response[len(m.Response)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipG2P(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthG2P
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
func (m *G2PLanguagesRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowG2P
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
			return fmt.Errorf("proto: G2PLanguagesRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: G2PLanguagesRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipG2P(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthG2P
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
func (m *G2PLanguagesResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowG2P
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
			return fmt.Errorf("proto: G2PLanguagesResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: G2PLanguagesResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Languages", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowG2P
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
				return ErrInvalidLengthG2P
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Languages = append(m.Languages, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Alphabets", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowG2P
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthG2P
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Alphabets = append(m.Alphabets, &G2PLanguagesResponse_Alphabet{})
			if err := m.Alphabets[len(m.Alphabets)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipG2P(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthG2P
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
func (m *G2PLanguagesResponse_Alphabet) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowG2P
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
			return fmt.Errorf("proto: Alphabet: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Alphabet: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Lang", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowG2P
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
				return ErrInvalidLengthG2P
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Lang = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HelpLink", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowG2P
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
				return ErrInvalidLengthG2P
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.HelpLink = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AllowedChars", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowG2P
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
				return ErrInvalidLengthG2P
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AllowedChars = append(m.AllowedChars, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipG2P(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthG2P
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
func skipG2P(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowG2P
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
					return 0, ErrIntOverflowG2P
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
					return 0, ErrIntOverflowG2P
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
				return 0, ErrInvalidLengthG2P
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowG2P
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
				next, err := skipG2P(dAtA[start:])
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
	ErrInvalidLengthG2P = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowG2P   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("ntx/v2t/engine/g2p.proto", fileDescriptorG2P) }

var fileDescriptorG2P = []byte{
	// 393 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x41, 0x4e, 0xdb, 0x40,
	0x14, 0x86, 0x3b, 0x49, 0x95, 0xda, 0x2f, 0x55, 0x55, 0x4d, 0x5b, 0xd5, 0x72, 0x2b, 0xcb, 0x72,
	0xbb, 0xc8, 0x06, 0x47, 0x1a, 0x58, 0xb2, 0x01, 0x84, 0x82, 0x94, 0x2c, 0x22, 0x83, 0x58, 0xb0,
	0x40, 0x9a, 0x84, 0x91, 0x6d, 0x61, 0xc6, 0x83, 0x3d, 0x09, 0x11, 0x27, 0xe1, 0x0c, 0x70, 0x11,
	0x96, 0x1c, 0x01, 0xc2, 0x45, 0x50, 0xc6, 0x63, 0x87, 0x80, 0x41, 0xac, 0xfc, 0xde, 0xf7, 0xfc,
	0xcf, 0xfc, 0xef, 0xd7, 0x80, 0xc5, 0xe5, 0xac, 0x3b, 0x25, 0xb2, 0xcb, 0x78, 0x18, 0x73, 0xd6,
	0x0d, 0x89, 0xf0, 0x45, 0x96, 0xca, 0x14, 0x7f, 0xe3, 0x72, 0xe6, 0x4f, 0x89, 0xf4, 0x8b, 0x89,
	0xb7, 0x07, 0xdf, 0x7b, 0x64, 0x78, 0x90, 0x51, 0x9e, 0x8f, 0xb3, 0x58, 0xc8, 0x38, 0xe5, 0xd8,
	0x06, 0x23, 0xcc, 0xa8, 0x88, 0xd8, 0x19, 0xb3, 0x90, 0x8b, 0x3a, 0x66, 0x50, 0xf5, 0xd8, 0x82,
	0x2f, 0x22, 0x4a, 0xf9, 0x62, 0xd4, 0x70, 0x9b, 0x1d, 0x33, 0x28, 0x5b, 0x6f, 0x03, 0xa0, 0x47,
	0x86, 0x01, 0x3b, 0x9f, 0xb0, 0x5c, 0x62, 0x0c, 0x9f, 0x13, 0xca, 0x43, 0xad, 0x57, 0xf5, 0x82,
	0x49, 0x36, 0x93, 0x56, 0xa3, 0x60, 0x8b, 0xda, 0xeb, 0x43, 0x5b, 0xa9, 0x72, 0x91, 0xf2, 0x9c,
	0xe1, 0x4d, 0x30, 0x32, 0x5d, 0x5b, 0xc8, 0x6d, 0x76, 0xda, 0xc4, 0xf5, 0x57, 0x1d, 0xfb, 0x2f,
	0xed, 0x06, 0x95, 0xc2, 0xfb, 0x05, 0x3f, 0x7a, 0x64, 0x38, 0xa0, 0x3c, 0x9c, 0xd0, 0x90, 0xe5,
	0xda, 0x8b, 0xf7, 0x80, 0xe0, 0xe7, 0x2a, 0xd7, 0xb7, 0xfd, 0x05, 0x33, 0x29, 0xa1, 0xba, 0xce,
	0x0c, 0x96, 0x00, 0xf7, 0xc1, 0xa4, 0x89, 0x88, 0xe8, 0x88, 0xc9, 0x5c, 0x2d, 0xdb, 0x26, 0x6b,
	0x35, 0x66, 0x5e, 0x1d, 0xeb, 0x6f, 0x69, 0x55, 0xb0, 0xd4, 0xdb, 0xc7, 0x60, 0x94, 0xb8, 0x36,
	0x1b, 0x1b, 0x8c, 0x88, 0x25, 0x62, 0x10, 0xf3, 0x53, 0x9d, 0x4f, 0xd5, 0x63, 0x0f, 0xbe, 0xd2,
	0x24, 0x49, 0x2f, 0xd8, 0xc9, 0x4e, 0x44, 0xb3, 0xdc, 0x6a, 0x2a, 0xa7, 0x2b, 0x8c, 0xdc, 0x20,
	0x15, 0xff, 0x3e, 0xcb, 0xa6, 0xf1, 0x98, 0xe1, 0x5d, 0x80, 0x32, 0xa4, 0x11, 0xc3, 0x76, 0x8d,
	0x6d, 0x1d, 0x8e, 0xfd, 0xa7, 0x76, 0xa6, 0x03, 0x3a, 0x04, 0xb3, 0x5a, 0x0f, 0xff, 0x7b, 0x7f,
	0xf9, 0xe2, 0xb8, 0xff, 0x1f, 0x49, 0x68, 0xbb, 0x7f, 0x3b, 0x77, 0xd0, 0xdd, 0xdc, 0x41, 0xf7,
	0x73, 0x07, 0x5d, 0x3d, 0x3a, 0x9f, 0xe0, 0xf7, 0xf8, 0x52, 0x29, 0xd5, 0x2b, 0x7d, 0xa6, 0x3f,
	0x6a, 0x15, 0xdf, 0xeb, 0xc6, 0x5b, 0x7f, 0x8c, 0x5a, 0x8a, 0xac, 0x3f, 0x05, 0x00, 0x00, 0xff,
	0xff, 0xc4, 0x32, 0xca, 0x93, 0xf5, 0x02, 0x00, 0x00,
}
