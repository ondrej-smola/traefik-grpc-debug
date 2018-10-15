package shared

import (
	"context"
	"sync"
	"time"

	"sync/atomic"

	"github.com/go-kit/kit/log"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
)

type (
	Runner struct {
		query CacheQuery
		next  scheduler.Runner

		queryInterval        time.Duration
		evictEveryNthAttempt int
		log                  log.Logger
	}

	RunnerOpt func(r *Runner)

	taskOrError struct {
		t    scheduler.Task
		err  error
		from string
	}
)

func WithRunnerQueryInterval(int time.Duration) RunnerOpt {
	return func(r *Runner) {
		r.queryInterval = int
	}
}

func WithRunnerEvictEveryNthAttempt(attempt int) RunnerOpt {
	return func(r *Runner) {
		r.evictEveryNthAttempt = attempt
	}
}

func WithRunnerLogger(log log.Logger) RunnerOpt {
	return func(r *Runner) {
		r.log = log
	}
}

func NewRunner(next scheduler.Runner, query CacheQuery, opts ...RunnerOpt) *Runner {
	r := &Runner{
		query:                query,
		next:                 next,
		queryInterval:        time.Second,
		log:                  log.NewNopLogger(),
		evictEveryNthAttempt: 3,
	}

	for _, o := range opts {
		o(r)
	}

	return r
}

func (c *Runner) Run(req scheduler.Request, ctx context.Context) (scheduler.Task, error) {
	doneLock := &sync.Mutex{}
	done := false

	nextCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	result := make(chan *taskOrError)

	// scheduler.TaskState
	scheduled := int32(scheduler.TaskState_WAITING)

	req = scheduler.RequestWithCallback(req, func(s scheduler.TaskState) {
		atomic.StoreInt32(&scheduled, int32(s))
	})

	go func() {
		t := time.NewTicker(c.queryInterval)
		defer t.Stop()

		attempt := 0

		for ; true; <-t.C {
			attempt += 1
			doneLock.Lock()
			cancel := done
			doneLock.Unlock()
			if cancel {
				return
			}

			req := &TaskRequest{
				Config: req.Config(),
				Evict:  attempt%c.evictEveryNthAttempt == 0 && scheduler.TaskState(atomic.LoadInt32(&scheduled)).IsWaiting(),
			}

			t, found, err := c.query.Get(req, nextCtx)
			if err != nil {
				c.log.Log("ev", "remote_cache_coord_get", "err", err)
			} else if found {
				doneLock.Lock()
				if done {
					c.log.Log("event", "releasing_remote_task", "cause", "request_already_completed", "task", t.Snapshot().Id)
					t.Release()
				} else {
					done = true
					result <- &taskOrError{
						t:    t,
						from: "remote_cache",
					}
				}
				doneLock.Unlock()
				return
			}
		}
	}()

	go func() {
		t, err := c.next.Run(req, nextCtx)
		doneLock.Lock()
		defer doneLock.Unlock()
		if !done {
			done = true
			result <- &taskOrError{
				t:    t,
				err:  err,
				from: "next_runner",
			}
		} else if err == nil {
			c.log.Log("event", "releasing_local_task", "cause", "request_already_completed", "task", t.Snapshot().Id)
			t.Release()
		}
	}()

	if res := <-result; res.err != nil {
		return nil, res.err
	} else {
		snap := res.t.Snapshot()
		req.OnStateChange(snap.State)
		c.log.Log("ev", "task_allocated", "from", res.from, "id", snap.Id, "endpoint", snap.Endpoint)
		return res.t, nil
	}
}
