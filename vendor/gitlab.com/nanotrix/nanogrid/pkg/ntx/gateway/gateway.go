package gateway

import (
	"io"
	"time"

	"github.com/gogo/protobuf/proto"

	"context"

	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/grpcutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/v2t/engine"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var ClientInactivityTimeout = errors.New("client inactivity timeout")
var ServerInactivityTimeout = errors.New("server inactivity timeout")

const (
	IdHeader        = "proxy-id"
	TaskIdHeader    = "task-id"
	RequestIdHeader = "request-id"
	SnapshotPath    = "/snapshot"
	ZkMembershipDir = "/ntx/gateway/members"
	ZkRoleDir       = "/ntx/gateway/roles"
)

type (
	Opt func(c *Gateway)

	MessageOffsetFn func(m *engine.EngineStream) (offset uint64, offsetSet bool)

	Gateway struct {
		id string

		inactivityTimeout       time.Duration
		serverInactivityTimeout time.Duration
		taskAllocationTimeout   time.Duration
		closeWaitTimeout        time.Duration
		maxAllocationAttempts   int

		runner       scheduler.Runner
		tokenDecoder auth.Decoder
		evs          ClientEvents

		msgOffsetFn MessageOffsetFn
		log         log.Logger
	}

	RequestMeta struct {
		GatewayId string
		TaskId    string
		RequestId string
	}

	RequestMetaNotification func(m RequestMeta)

	Stream interface {
		RecvMsg() (*engine.EngineStream, error)
		SendMsg(*engine.EngineStream) error
	}

	Src interface {
		Stream
		// headers are sent with first SendMsg call, ignored afterwards
		SendHeaders(headers map[string]string) error
	}

	Dest interface {
		Stream
		Close() error
	}

	Doer interface {
		Do(reqId string, from Src, to DestFn, tkn *auth.NtxToken, ctx context.Context) error
	}

	DoFn func(reqId string, from Src, to DestFn, tkn *auth.NtxToken, ctx context.Context) error

	DestFn func(conn *grpc.ClientConn) (Dest, error)
)

func (f DoFn) Do(reqId string, from Src, to DestFn, tkn *auth.NtxToken, ctx context.Context) error {
	return f(reqId, from, to, tkn, ctx)
}

func WithInactivityTimeout(t time.Duration) Opt {
	return func(c *Gateway) {
		c.inactivityTimeout = t
	}
}

func WithCloseWaitTimeout(t time.Duration) Opt {
	return func(c *Gateway) {
		c.closeWaitTimeout = t
	}
}

func WithTaskAllocationTimeout(t time.Duration) Opt {
	return func(c *Gateway) {
		c.taskAllocationTimeout = t
	}
}

func WithMaxAllocationAttempts(count int) Opt {
	if count < 1 {
		panic(fmt.Errorf("count must be >= 1, got %v", count))
	}
	return func(c *Gateway) {
		c.maxAllocationAttempts = count
	}
}

func WithLogger(log log.Logger) Opt {
	return func(c *Gateway) {
		c.log = log
	}
}

func WithId(id string) Opt {
	return func(c *Gateway) {
		c.id = id
	}
}

func WithClientEvents(ev ClientEvents) Opt {
	return func(c *Gateway) {
		c.evs = ev
	}
}

func WithMsgOffsetFn(fn MessageOffsetFn) Opt {
	return func(c *Gateway) {
		c.msgOffsetFn = fn
	}
}

func NewGateway(tasks scheduler.Runner, opts ...Opt) *Gateway {
	p := &Gateway{
		id:                    "default",
		runner:                tasks,
		evs:                   NewClientEventsStore(),
		inactivityTimeout:     30 * time.Second,
		taskAllocationTimeout: 15 * time.Minute,
		closeWaitTimeout:      3 * time.Second,
		log:                   log.NewNopLogger(),
		maxAllocationAttempts: 1,
		msgOffsetFn: func(*engine.EngineStream) (uint64, bool) {
			return 0, false
		},
	}

	for _, o := range opts {
		o(p)
	}
	return p
}

func (s *Gateway) Do(reqId string, from Src, to DestFn, tkn *auth.NtxToken, ctx context.Context) error {
	s.log.Log("event", "request_start", "req_id", reqId, "debug", true)

	err := s.do(reqId, from, to, tkn, ctx)

	if err != nil {
		s.log.Log("event", "request_failed", "req_id", reqId, "err", err)

	} else {
		s.log.Log("event", "request_done", "req_id", reqId, "debug", true)
	}

	return err
}

func (s *Gateway) do(reqId string, from Src, to DestFn, tkn *auth.NtxToken, ctx context.Context) (err error) {
	s.evs.New(reqId, tkn)
	defer func() {
		if err != nil {
			s.evs.Completed(reqId, err.Error())
		} else {
			s.evs.Completed(reqId)
		}
	}()

	t, err := s.allocateNewTask(tkn.Task, ctx)
	if err != nil {
		return err
	}

	defer t.Release()
	taskSnap := t.Snapshot()
	s.log.Log("ev", "task_running", "id", taskSnap.Id, "debug", true)
	s.evs.Scheduled(reqId, taskSnap)

	if err := from.SendHeaders(
		map[string]string{
			IdHeader:        s.id,
			TaskIdHeader:    taskSnap.Id,
			RequestIdHeader: reqId}); err != nil {

		return errors.Wrap(err, "header: send")
	}

	return errors.Wrapf(s.proxy(from, to, reqId, taskSnap.Endpoint, ctx), "do failed, task '%v', endpoint '%v'", taskSnap.Id, taskSnap.Endpoint)
}

func (s *Gateway) allocateNewTask(req *auth.NtxToken_Task, ctx context.Context) (scheduler.Task, error) {
	runReq := scheduler.NewRequest(req.Cfg)

	ctx, cancel := context.WithTimeout(ctx, s.taskAllocationTimeout)
	defer cancel()

	var lastErr error

	for i := 0; i < s.maxAllocationAttempts; i++ {
		t, err := s.runner.Run(runReq, ctx)
		if err != nil {
			lastErr = err
			s.log.Log("ev", "task_allocation_failed", "err", err, "attempt", 1)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		} else {
			return t, nil
		}
	}

	return nil, errors.Wrap(lastErr, "all allocation attempts failed")
}

func (s *Gateway) proxy(from Src, makeDest DestFn, reqId, endpoint string, ctx context.Context) error {
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		return status.Errorf(codes.Internal, "failed to connect to proxied task")
	}

	defer conn.Close()

	to, err := makeDest(conn)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to call allocated egress: %v", err)
	}

	ingress := make(chan error, 1)
	egress := make(chan error, 1)

	s.log.Log("event", "do_proxy", "to", endpoint, "debug", true)

	clientInactivityTimeout := time.NewTimer(s.inactivityTimeout)
	serverInactivityTimeout := time.NewTimer(s.inactivityTimeout)

	// ingress
	// client -> proxy -> engine
	go func() {
		defer close(ingress)
		for {

			msg, err := from.RecvMsg()
			if err != nil {
				if e := to.Close(); e != nil {
					s.log.Log("event", "ingress_close_failed", "err", e, "debug", true)
				}

				if err != io.EOF && !grpcutil.IsContextCancelled(err) {
					ingress <- errors.Wrap(err, "Ingress: recv")
				}
				return
			}

			clientInactivityTimeout.Reset(s.inactivityTimeout)

			s.evs.MsgIn(reqId, uint64(proto.Size(msg)))

			err = to.SendMsg(msg)
			if err != nil {
				ingress <- errors.Wrap(err, "ingress: send")
				return
			}
		}
	}()

	// egress
	// engine -> proxy -> client
	go func() {
		defer close(egress)
		firstRecv := true

		for {
			msg, err := to.RecvMsg()
			if err != nil {
				if err != io.EOF && !grpcutil.IsContextCancelled(err) {
					egress <- errors.Wrap(err, "egress: recv")
				}
				return
			}
			serverInactivityTimeout.Reset(s.inactivityTimeout)

			if firstRecv {
				firstRecv = false
				s.evs.Started(reqId)
			}

			err = from.SendMsg(msg)
			if err != nil {
				egress <- errors.Wrap(err, "egress: send")
				return
			}

			if offset, ok := s.msgOffsetFn(msg); ok {
				s.evs.StreamLength(reqId, offset)
			}

			s.evs.MsgOut(reqId, uint64(proto.Size(msg)))
		}
	}()

	var ingressErr, egressErr error

	select {
	case <-clientInactivityTimeout.C:
		return ClientInactivityTimeout
	case <-serverInactivityTimeout.C:
		return ServerInactivityTimeout
	case ingressErr = <-ingress:
		if ingressErr == nil {
			select {
			case egressErr = <-egress:
			case <-time.After(s.closeWaitTimeout):
				egressErr = errors.New("egress: close timeout")
			}
		}
	case egressErr = <-egress:
		if egressErr == nil {
			select {
			case ingressErr = <-ingress:
			case <-time.After(s.closeWaitTimeout):
				egressErr = errors.New("ingress: close timeout")
			}
		}
	}

	if egressErr == nil && ingressErr == nil {
		return nil
	} else if ingressErr == nil {
		return egressErr
	} else if egressErr == nil {
		return ingressErr
	} else {
		return errors.Errorf("'%v' and '%v'", ingressErr, egressErr)
	}
}

func GatewayRequestMetaFromMetadata(md metadata.MD) RequestMeta {
	return RequestMeta{
		GatewayId: grpcutil.GetFirstFromMD(md, IdHeader),
		RequestId: grpcutil.GetFirstFromMD(md, RequestIdHeader),
		TaskId:    grpcutil.GetFirstFromMD(md, TaskIdHeader),
	}
}

func (m *RequestMeta) String() string {
	return fmt.Sprintf("RequestMeta[gw=%v,req=%v,task=%v]", m.GatewayId, m.RequestId, m.TaskId)
}
