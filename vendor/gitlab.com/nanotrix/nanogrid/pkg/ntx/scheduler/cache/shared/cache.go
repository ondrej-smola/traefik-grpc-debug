package shared

import (
	"context"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
)

type (
	CacheQuery interface {
		Get(req *TaskRequest, ctx context.Context) (scheduler.Task, bool, error)
	}

	CacheQueryFn func(req *TaskRequest, ctx context.Context) (scheduler.Task, bool, error)

	Cache interface {
		CacheQuery
		Ping(ctx context.Context) error
		Destroy(ctx context.Context) error
		Snapshot() []*scheduler.TaskSnapshot
	}

	CacheProvider func(endpoint string) (Cache, error)

	taskCache struct {
		cl      TaskSharingServiceClient
		timeout time.Duration

		log log.Logger

		sync.Mutex
		tasks map[string]scheduler.Task

		onDestroy context.CancelFunc
	}

	remoteTask struct {
		owner      *taskCache
		returnOnce *sync.Once

		// mutex for
		sync.Mutex
		snap *scheduler.TaskSnapshot
	}

	CacheOpt func(r *taskCache)
)

func (f CacheQueryFn) Get(req *TaskRequest, ctx context.Context) (scheduler.Task, bool, error) {
	return f(req, ctx)
}

func WithRemoteCacheLogger(l log.Logger) CacheOpt {
	return func(r *taskCache) {
		r.log = l
	}
}

func WithRemoteCacheOnDestroy(on context.CancelFunc) CacheOpt {
	return func(r *taskCache) {
		r.onDestroy = on
	}
}

func NewRemoteCache(cl TaskSharingServiceClient, opts ...CacheOpt) *taskCache {
	r := &taskCache{
		cl:        cl,
		tasks:     map[string]scheduler.Task{},
		log:       log.NewNopLogger(),
		timeout:   3 * time.Second,
		onDestroy: func() {},
	}

	for _, o := range opts {
		o(r)
	}

	return r
}

func (r *taskCache) Destroy(ctx context.Context) error {
	r.Lock()
	taskIds := r.unsafeTaskIdsCopy()
	r.tasks = map[string]scheduler.Task{}
	r.Unlock()

	_, err := r.cl.Kill(ctx, &Tasks{Ids: taskIds})
	r.onDestroy()

	return errors.Wrap(err, "destroy: kill tasks")
}

func (r *taskCache) unsafeTaskIdsCopy() []string {
	cpy := make([]string, len(r.tasks))
	i := 0
	for k := range r.tasks {
		cpy[i] = k
		i++
	}
	return cpy
}

func (r *taskCache) Ping(ctx context.Context) error {
	r.Lock()
	cpy := r.unsafeTaskIdsCopy()
	r.Unlock()
	if len(cpy) > 0 {
		_, err := r.cl.Ping(ctx, &Tasks{Ids: cpy})
		return errors.Wrap(err, "ping")
	} else {
		return nil
	}
}

func (r *taskCache) Get(req *TaskRequest, ctx context.Context) (scheduler.Task, bool, error) {
	if resp, err := r.cl.Request(ctx, req); err != nil {
		return nil, false, errors.Wrap(err, "request")
	} else if err = resp.Valid(); err != nil {
		return nil, false, errors.Wrap(err, "response")
	} else if resp.Found {

		snap := &scheduler.TaskSnapshot{
			Id:       resp.Id,
			Endpoint: resp.Endpoint,
			Owner:    resp.Owner,
			Config:   req.Config.Clone(),
			State:    scheduler.TaskState_RUNNING,
			Created:  uint64(timeutil.NowUTC().Unix()),
		}

		remTask := &remoteTask{
			snap:       snap,
			owner:      r,
			returnOnce: &sync.Once{},
		}

		r.Lock()
		r.tasks[resp.Id] = remTask
		r.Unlock()

		return remTask, true, nil
	} else {
		return nil, false, nil
	}
}

func (c *taskCache) Snapshot() []*scheduler.TaskSnapshot {
	c.Lock()
	defer c.Unlock()

	res := make([]*scheduler.TaskSnapshot, len(c.tasks))
	i := 0
	for _, t := range c.tasks {
		res[i] = t.Snapshot()
		i++
	}

	return res
}

func (r *taskCache) release(t *remoteTask) {
	r.Lock()
	tid := t.snap.Id

	if _, ok := r.tasks[tid]; !ok {
		r.log.Log("ev", "release_unknown_task", "tid", tid)
		r.Unlock()
		return
	}
	delete(r.tasks, tid)
	r.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		defer cancel()

		if _, err := r.cl.Release(ctx, &Tasks{Ids: []string{tid}}); err != nil {
			r.log.Log("ev", "failed_to_release", "tid", tid, "err", err)
		} else {
			r.log.Log("ev", "released_task", "tid", tid, "debug", true)
		}
	}()
}

func (r *taskCache) kill(t *remoteTask) {
	r.Lock()
	tid := t.snap.Id
	if _, ok := r.tasks[tid]; !ok {
		r.log.Log("ev", "kill_unknown_task", "tid", tid)
		r.Unlock()
		return
	}
	delete(r.tasks, tid)
	r.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		defer cancel()

		if _, err := r.cl.Kill(ctx, &Tasks{Ids: []string{tid}}); err != nil {
			r.log.Log("ev", "failed_to_kill", "tid", tid, "err", err)
		} else {
			r.log.Log("ev", "killed_task", "tid", tid, "debug", true)
		}
	}()

}

func (c *remoteTask) Snapshot() *scheduler.TaskSnapshot {
	c.Lock()
	defer c.Unlock()
	cpy := *c.snap
	return &cpy
}

func (c *remoteTask) Release() {
	c.Lock()
	c.snap.State = scheduler.TaskState_KILLED
	c.Unlock()
	c.returnOnce.Do(func() {
		c.owner.release(c)
	})
}

func (c *remoteTask) Kill() {
	c.Lock()
	c.snap.State = scheduler.TaskState_KILLED
	c.Unlock()
	c.returnOnce.Do(func() {
		c.owner.kill(c)
	})
}
