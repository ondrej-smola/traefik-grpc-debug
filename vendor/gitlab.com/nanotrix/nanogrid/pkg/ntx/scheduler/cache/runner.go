package cache

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"sync"
	"time"

	"sync/atomic"

	"github.com/go-kit/kit/log"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type (
	Runner struct {
		cacheCheckInterval       time.Duration
		releaseResourcesInterval time.Duration
		evictResources           bool

		log log.Logger

		parent scheduler.Runner
		cache  TaskCache
	}

	RunnerOpt func(r *Runner)

	runnerTask struct {
		req   scheduler.Request
		prev  scheduler.Task
		owner *Runner

		cacheKey string

		releaseOnce sync.Once
	}

	taskOrError struct {
		t    scheduler.Task
		err  error
		from string
	}
)

func WithRunnerCacheCheckInterval(int time.Duration) RunnerOpt {
	return func(r *Runner) {
		r.cacheCheckInterval = int
	}
}

func WithRunnerEvictResources(evict bool) RunnerOpt {
	return func(r *Runner) {
		r.evictResources = evict
	}
}

func WithRunnerResourceReleaseInterval(int time.Duration) RunnerOpt {
	return func(r *Runner) {
		r.releaseResourcesInterval = int
	}
}

func WithRunnerLogger(log log.Logger) RunnerOpt {
	return func(r *Runner) {
		r.log = log
	}
}

func WithRunnerCache(c TaskCache) RunnerOpt {
	return func(r *Runner) {
		r.cache = c
	}
}

func NewRunner(cacheFor scheduler.Runner, opts ...RunnerOpt) *Runner {
	r := &Runner{
		cacheCheckInterval:       50 * time.Millisecond,
		releaseResourcesInterval: 3 * time.Second,
		log:                      log.NewNopLogger(),
		parent:                   cacheFor,
		cache:                    NewNoopCache(),
	}

	for _, o := range opts {
		o(r)
	}

	return r
}

func TaskConfigurationToCacheKey(cfg *scheduler.TaskConfiguration) string {
	sum := sha1.Sum(util.MustJsonPb(cfg))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func (c *Runner) Run(req scheduler.Request, ctx context.Context) (scheduler.Task, error) {
	key := TaskConfigurationToCacheKey(req.Config())

	// first check for cached task
	if t, found := c.cache.Get(key); found {
		snap := t.Snapshot()
		c.log.Log("ev", "task_allocated", "from", "cache", "id", snap.Id, "endpoint", snap.Endpoint)
		return &runnerTask{req: req, prev: t, owner: c, cacheKey: key, releaseOnce: sync.Once{}}, nil
	}

	lock := sync.Mutex{}
	done := false
	// scheduler.TaskState
	scheduled := int32(scheduler.TaskState_WAITING)

	result := make(chan *taskOrError)

	newReq := scheduler.RequestWithCallback(req, func(s scheduler.TaskState) {
		atomic.StoreInt32(&scheduled, int32(s))
	})

	allocCtx, allocCancel := context.WithCancel(ctx)

	// allocate using parent Runner
	go func() {
		t, err := c.parent.Run(newReq, allocCtx)
		lock.Lock()
		if !done {
			done = true
			result <- &taskOrError{t: t, err: err, from: "next_runner"}
		} else if err == nil {
			c.log.Log("event", "caching_task", "cause", "request_already_completed", "tid", t.Snapshot().Id)
			c.cache.Add(key, t)
		}
		lock.Unlock()
	}()

	// periodically check cache if there is new cached task that can be used instead of the currently allocating
	go func() {
		ticker := time.NewTicker(c.cacheCheckInterval)
		defer ticker.Stop()
		for range ticker.C {
			lock.Lock()
			if done {
				lock.Unlock()
				return
			}

			t, found := c.cache.Get(key)
			if found {
				allocCancel()
				done = true
				result <- &taskOrError{t: t, from: "cache"}
			}
			lock.Unlock()
			if found {
				return
			}
		}
	}()

	if c.evictResources {
		// try to release resources when there is no notification from parent that task has been at least scheduled
		go func() {
			ticker := time.NewTicker(c.releaseResourcesInterval)
			defer ticker.Stop()
			for range ticker.C {
				lock.Lock()
				if done || !scheduler.TaskState(atomic.LoadInt32(&scheduled)).IsWaiting() {
					lock.Unlock()
					return
				}
				lock.Unlock()

				resources := ResourcesFromTaskConfig(req.Config())

				if c.cache.Evict(resources.Clone()) {
					c.log.Log("ev", "released_resources_from_cache", "res", resources, "cause", "task_waiting")
				} else {
					c.log.Log("ev", "unable_to_release_resources_from_cache", "res", resources)
				}
			}
		}()
	}

	if res := <-result; res.err != nil {
		return nil, res.err
	} else {
		snap := res.t.Snapshot()
		c.log.Log("ev", "task_allocated", "from", res.from, "id", snap.Id, "endpoint", snap.Endpoint)
		return &runnerTask{req: req, prev: res.t, owner: c, cacheKey: key, releaseOnce: sync.Once{}}, nil
	}
}

func (c *Runner) releaseTask(t *runnerTask) {
	snap := t.Snapshot()

	if snap.State.IsRunning() {
		c.log.Log("ev", "caching_task", "id", snap.Id, "key", t.cacheKey, "debug", true)
		c.cache.Add(t.cacheKey, t.prev)
	} else {
		c.log.Log("ev", "killing_task", "cause", "not_running", "id", snap.Id, "debug", true)
		t.prev.Kill()
	}
}

func (c *runnerTask) Snapshot() *scheduler.TaskSnapshot {
	return c.prev.Snapshot()
}

func (c *runnerTask) Release() {
	c.releaseOnce.Do(func() {
		c.owner.releaseTask(c)
	})
}

func (c *runnerTask) Kill() {
	c.prev.Kill()
}
