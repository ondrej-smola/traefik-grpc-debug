package shared

import (
	"context"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
	xcontext "golang.org/x/net/context"
	"google.golang.org/grpc"
)

type (
	TestCacheGetRequest struct {
		Req *TaskRequest
		Ctx context.Context
	}

	TestCacheGetResponse struct {
		T     scheduler.Task
		Found bool
		Err   error
	}

	TestCache struct {
		GetIn   chan *TestCacheGetRequest
		GetOut  chan *TestCacheGetResponse
		PingIn  chan context.Context
		PingOut chan error

		DestroyIn  chan context.Context
		DestroyOut chan error

		SnapshotIn  chan bool
		SnapshotOut chan []*scheduler.TaskSnapshot
	}

	TestTaskResponseWithError struct {
		Resp *TaskResponse
		Err  error
	}

	TestTaskSharingServiceClient struct {
		PingIn, ReleaseIn, KillIn    chan *Tasks
		PingOut, ReleaseOut, KillOut chan error

		RequestIn  chan *TaskRequest
		RequestOut chan *TestTaskResponseWithError
	}
)

func NewTestCache() *TestCache {
	return &TestCache{
		GetIn:  make(chan *TestCacheGetRequest),
		GetOut: make(chan *TestCacheGetResponse),

		PingIn:  make(chan context.Context),
		PingOut: make(chan error),

		DestroyIn:  make(chan context.Context),
		DestroyOut: make(chan error),

		SnapshotIn:  make(chan bool),
		SnapshotOut: make(chan []*scheduler.TaskSnapshot),
	}
}

func (t *TestCache) Get(req *TaskRequest, ctx context.Context) (scheduler.Task, bool, error) {
	t.GetIn <- &TestCacheGetRequest{
		Req: req,
		Ctx: ctx,
	}

	resp := <-t.GetOut
	return resp.T, resp.Found, resp.Err
}

func (t *TestCache) Ping(ctx context.Context) error {
	t.PingIn <- ctx
	return <-t.PingOut
}

func (t *TestCache) Destroy(ctx context.Context) error {
	t.DestroyIn <- ctx
	return <-t.DestroyOut
}

func (t *TestCache) Snapshot() []*scheduler.TaskSnapshot {
	t.SnapshotIn <- true
	return <-t.SnapshotOut
}

func NewTestTaskSharingServiceClient() *TestTaskSharingServiceClient {
	return &TestTaskSharingServiceClient{
		PingIn:     make(chan *Tasks),
		PingOut:    make(chan error),
		ReleaseIn:  make(chan *Tasks),
		ReleaseOut: make(chan error),
		KillIn:     make(chan *Tasks),
		KillOut:    make(chan error),
		RequestIn:  make(chan *TaskRequest),
		RequestOut: make(chan *TestTaskResponseWithError),
	}
}

func (c *TestTaskSharingServiceClient) Ping(ctx xcontext.Context, in *Tasks, opts ...grpc.CallOption) (*ntx.Empty, error) {
	c.PingIn <- in
	return &ntx.Empty{}, <-c.PingOut
}
func (c *TestTaskSharingServiceClient) Release(ctx xcontext.Context, in *Tasks, opts ...grpc.CallOption) (*ntx.Empty, error) {
	c.ReleaseIn <- in
	return &ntx.Empty{}, <-c.ReleaseOut
}
func (c *TestTaskSharingServiceClient) Kill(ctx xcontext.Context, in *Tasks, opts ...grpc.CallOption) (*ntx.Empty, error) {
	c.KillIn <- in
	return &ntx.Empty{}, <-c.KillOut
}
func (c *TestTaskSharingServiceClient) Request(ctx xcontext.Context, in *TaskRequest, opts ...grpc.CallOption) (*TaskResponse, error) {
	c.RequestIn <- in
	out := <-c.RequestOut
	return out.Resp, out.Err
}
