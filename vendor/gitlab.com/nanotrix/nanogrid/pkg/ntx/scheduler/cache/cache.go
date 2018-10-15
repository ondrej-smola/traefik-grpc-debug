package cache

import (
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
)

type (
	TaskQuery interface {
		Get(key string) (t scheduler.Task, found bool)
	}

	TaskGetFn func(key string) (t scheduler.Task, found bool)

	Evictor interface {
		Evict(resources *Resources) bool
	}

	// thread safe
	TaskCache interface {
		TaskQuery
		Add(key string, t scheduler.Task)
		Snapshot() []*scheduler.TaskSnapshot
		Destroy()
		GC()
		Evictor
	}

	DefaultTaskCache struct {
		cacheLineFn LineFn
		lines       map[string]Line
		sync.Mutex

		log log.Logger
	}

	LineFn func() Line

	TaskCacheOpt func(p *DefaultTaskCache)
)

func (f TaskGetFn) Get(key string) (t scheduler.Task, found bool) {
	return f(key)
}

func WithDefaultTaskCacheLogger(log log.Logger) TaskCacheOpt {
	return func(p *DefaultTaskCache) {
		p.log = log
	}
}

func New(lineFn LineFn, opts ...TaskCacheOpt) *DefaultTaskCache {
	r := &DefaultTaskCache{
		cacheLineFn: lineFn,
		lines:       make(map[string]Line),
		log:         log.NewNopLogger(),
	}

	for _, o := range opts {
		o(r)
	}

	return r
}

func (e *DefaultTaskCache) Get(key string) (scheduler.Task, bool) {
	e.Lock()
	defer e.Unlock()

	r, found := e.lines[key]
	if !found {
		return nil, false
	}

	return r.Get()
}

func (e *DefaultTaskCache) Add(key string, t scheduler.Task) {
	e.Lock()
	defer e.Unlock()

	l, found := e.lines[key]
	if !found {
		l = e.cacheLineFn()
		e.lines[key] = l
	}

	l.Add(t)
}

func (e *DefaultTaskCache) Size() int {
	e.Lock()
	total := 0
	for _, r := range e.lines {
		total += r.Size()
	}
	e.Unlock()
	return total
}

func (e *DefaultTaskCache) Snapshot() []*scheduler.TaskSnapshot {
	e.Lock()
	snapshots := []*scheduler.TaskSnapshot{}
	for _, r := range e.lines {
		snapshots = append(snapshots, r.Snapshot()...)
	}
	e.Unlock()
	return snapshots
}

func (e *DefaultTaskCache) GC() {
	e.Lock()
	for k, r := range e.lines {
		r.GC()

		if r.Size() == 0 {
			r.Destroy()
			delete(e.lines, k)
		}
	}
	e.Unlock()
}

func (e *DefaultTaskCache) Destroy() {
	e.Lock()
	for _, r := range e.lines {
		r.Destroy()
	}
	e.lines = make(map[string]Line)
	e.Unlock()
}

func (e *DefaultTaskCache) Evict(r *Resources) bool {
	if err := r.Valid(); err != nil {
		panic(errors.Wrap(err, "Invalid argument: Resources"))
	}

	remCpu := r.Cpus
	remMem := r.Mem

	if remCpu <= 0 && remMem <= 0 {
		return true
	}

	e.Lock()
	defer e.Unlock()

	for k, v := range e.lines {
		if remCpu <= 0 && remMem <= 0 {
			break
		}

		for v.Size() > 0 {
			t, _ := v.Get()
			t.Release()

			cfg := t.Snapshot().Config

			remCpu -= cfg.Cpus
			remMem -= cfg.Mem

			e.log.Log(
				"ev", "release_task",
				"tid", t.Snapshot().Id,
				"cause", "evict",
				"task_cpu", cfg.Cpus,
				"task_mem", cfg.Mem,
				"rem_cpu", remCpu,
				"rem_mem", remMem,
			)

			if remCpu <= 0 && remMem <= 0 {
				break
			}
		}

		if v.Size() == 0 {
			delete(e.lines, k)
		}
	}

	return remCpu <= 0 && remMem <= 0
}
