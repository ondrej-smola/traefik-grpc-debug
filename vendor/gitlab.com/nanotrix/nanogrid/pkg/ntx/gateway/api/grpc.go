package api

import (
	"strconv"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/gateway"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/grpcutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/middleware"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/v2t/engine"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	Grpc struct {
		do    gateway.DoFn
		tknFn middleware.GrpcNtxTokenFn
	}

	grpcSrc struct {
		engine.EngineService_StreamingRecognizeServer
	}

	grpcDest struct {
		engine.EngineService_StreamingRecognizeClient
	}
)

func NewGrpc(p gateway.DoFn, tknFn middleware.GrpcNtxTokenFn) *Grpc {
	return &Grpc{do: p, tknFn: tknFn}
}

func (s *grpcSrc) RecvMsg() (*engine.EngineStream, error) {
	str, err := s.Recv()
	if err != nil {
		return nil, err
	} else {
		return str, str.Valid()
	}
}

func (s *grpcSrc) SendMsg(m *engine.EngineStream) error {
	return s.Send(m)
}

func (s *grpcSrc) SendHeaders(headers map[string]string) error {
	return s.SendHeader(metadata.New(headers))
}

func (p *grpcDest) RecvMsg() (*engine.EngineStream, error) {
	return p.Recv()
}

func (p *grpcDest) SendMsg(m *engine.EngineStream) error {
	return p.Send(m)
}

func (p *grpcDest) Close() error {
	return p.CloseSend()
}

func (s *Grpc) isFlowControlDisabled(md metadata.MD) bool {
	ok, _ := strconv.ParseBool(grpcutil.GetFirstFromMD(md, "no-flow-control"))
	return ok
}

func (s *Grpc) StreamingRecognize(serv engine.EngineService_StreamingRecognizeServer) error {
	md, ok := metadata.FromIncomingContext(serv.Context())
	if !ok {
		return status.Error(codes.InvalidArgument, "no header set")
	}

	tkn, err := s.tknFn(md)
	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	if tkn.Task == nil {
		return status.Error(codes.Unauthenticated, auth.ErrMissingNtxTokenTask.Error())
	}

	src := gateway.Src(&grpcSrc{serv})

	if s.isFlowControlDisabled(md) {
		src = NewPushOnlySrc(src, serv.Context())
	}

	err = s.do(util.NewUUID(), src, func(conn *grpc.ClientConn) (gateway.Dest, error) {
		cl, err := engine.NewEngineServiceClient(conn).StreamingRecognize(serv.Context())
		return &grpcDest{cl}, err
	}, tkn, serv.Context())

	if gateway.IsLimitError(err) {
		return status.Error(codes.ResourceExhausted, err.Error())
	} else {
		return grpcutil.ToGRPCError(err)
	}
}
