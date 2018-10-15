package gateway

import (
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/grpcutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/hello"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/middleware"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/v2t/engine"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGatewaySnapshotClient() *Client {
	return &Client{
		Id:           util.NewUUID(),
		State:        Client_RUNNING,
		Token:        auth.TestNtxToken(),
		TaskId:       util.NewUUID(),
		TaskEndpoint: "localhost:5050",
		Created:      uint64(timeutil.NowUTC().Unix()),
		Scheduled:    uint64(timeutil.NowUTC().Add(time.Second).Unix()),
		Started:      uint64(timeutil.NowUTC().Add(2 * time.Second).Unix()),
		Completed:    uint64(timeutil.NowUTC().Add(time.Minute).Unix()),
		Seqn:         rand.Uint64(),
		MsgsIn:       500,
		MsgsOut:      500,
		BytesOut:     1e6,
		BytesIn:      1e6,
	}
}

type TestProxyDestServer struct {
	StreamIn       chan engine.EngineService_StreamingRecognizeServer
	StreamOut      chan error
	HealthCheckIn  chan *ntx.ServiceNameWithContext
	HealthCheckOut chan *ntx.ServingStatusOrError
}

func (t *TestProxyDestServer) StreamingRecognize(src engine.EngineService_StreamingRecognizeServer) error {
	t.StreamIn <- src
	return <-t.StreamOut
}

func (s *TestProxyDestServer) Check(ctx context.Context, req *ntx.HealthCheckRequest) (*ntx.HealthCheckResponse, error) {
	s.HealthCheckIn <- &ntx.ServiceNameWithContext{Service: req.Service, Context: ctx}

	resp := <-s.HealthCheckOut
	return &ntx.HealthCheckResponse{Status: resp.Status}, resp.Err
}

func (t *TestProxyDestServer) SayHello(ctx context.Context, h *hello.Hello) (*hello.Hello, error) {
	return nil, errors.New("Not implemented")
}

func (t *TestProxyDestServer) SayHelloNTimes(*hello.Hello, hello.HelloService_SayHelloNTimesServer) error {
	return errors.New("Not implemented")
}

func NewTestProxyDestServer() *TestProxyDestServer {
	return &TestProxyDestServer{
		StreamIn:       make(chan engine.EngineService_StreamingRecognizeServer),
		StreamOut:      make(chan error),
		HealthCheckIn:  make(chan *ntx.ServiceNameWithContext),
		HealthCheckOut: make(chan *ntx.ServingStatusOrError),
	}
}

type (
	TestProxyServer struct {
		p     DoFn
		tknFn middleware.GrpcNtxTokenFn
	}

	testProxySrc struct {
		engine.EngineService_StreamingRecognizeServer
	}

	testProxyDest struct {
		engine.EngineService_StreamingRecognizeClient
	}
)

func NewTestProxyServer(p DoFn, tknFn middleware.GrpcNtxTokenFn) *TestProxyServer {
	return &TestProxyServer{p: p, tknFn: tknFn}
}

func (s *testProxySrc) RecvMsg() (*engine.EngineStream, error) {
	return s.Recv()
}

func (s *testProxySrc) SendMsg(m *engine.EngineStream) error {
	return s.Send(m)
}

func (s *testProxySrc) SendHeaders(headers map[string]string) error {
	return s.SendHeader(metadata.New(headers))
}

func (p *testProxyDest) RecvMsg() (*engine.EngineStream, error) {
	return p.Recv()
}

func (p *testProxyDest) SendMsg(m *engine.EngineStream) error {
	return p.Send(m)
}

func (p *testProxyDest) Close() error {
	return p.CloseSend()
}

func (t *TestProxyServer) StreamingRecognize(src engine.EngineService_StreamingRecognizeServer) error {
	md, ok := metadata.FromIncomingContext(src.Context())
	if !ok {
		return status.Error(codes.ResourceExhausted, "no header set")
	}

	tkn, err := t.tknFn(md)

	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	err = t.p(util.NewUUID(), &testProxySrc{src}, func(conn *grpc.ClientConn) (Dest, error) {
		ctx := src.Context()

		incomingMD, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = metadata.NewOutgoingContext(ctx, incomingMD)
		}

		cl, err := engine.NewEngineServiceClient(conn).StreamingRecognize(ctx)
		return &testProxyDest{cl}, err
	}, tkn, src.Context())

	return grpcutil.ToGRPCError(err)
}
