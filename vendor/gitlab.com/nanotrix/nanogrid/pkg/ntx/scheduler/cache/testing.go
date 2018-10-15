package cache

import (
	"sync/atomic"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
)

type TestingTaskCacheMetrics struct {
	Cached int32
}

func (p *TestingTaskCacheMetrics) IncCachedTaskCount() {
	atomic.AddInt32(&p.Cached, 1)
}

func (p *TestingTaskCacheMetrics) DecCachedTaskCount() {
	atomic.AddInt32(&p.Cached, -1)
}

type NoopCache struct {
	DoNotDeleteTask bool // only for testing
}

func NewNoopCache() *NoopCache {
	return &NoopCache{}
}

func (e *NoopCache) Get(key string) (scheduler.Task, bool) {
	return nil, false
}

func (e *NoopCache) Add(key string, t scheduler.Task) {
	if !e.DoNotDeleteTask {
		t.Release()
	}
}

func (e *NoopCache) TryRelease(r *Resources) bool {
	return false
}

func (e *NoopCache) Destroy() {}
func (e *NoopCache) Size() int {
	return 0
}

func (e *NoopCache) GC() {}
func (e *NoopCache) Snapshot() []*scheduler.TaskSnapshot {
	return nil
}

func (e *NoopCache) Evict(resources *Resources) bool {
	return false
}

type TestTaskWithKey struct {
	Key  string
	Task scheduler.Task
}

type GetAddTestCache struct {
	AddIn  chan *TestTaskWithKey
	AddOut chan bool
	GetIn  chan string
	GetOut chan scheduler.Task
}

func NewGetAddTestCache() *GetAddTestCache {
	return &GetAddTestCache{
		AddIn:  make(chan *TestTaskWithKey),
		AddOut: make(chan bool),
		GetIn:  make(chan string),
		GetOut: make(chan scheduler.Task),
	}
}

func (e *GetAddTestCache) Get(key string) (scheduler.Task, bool) {
	e.GetIn <- key
	res := <-e.GetOut
	return res, res != nil
}

func (e *GetAddTestCache) Add(key string, t scheduler.Task) {
	e.AddIn <- &TestTaskWithKey{Key: key, Task: t}
	<-e.AddOut
}

func (e *GetAddTestCache) TryRelease(r *Resources) bool {
	return false
}

func (e *GetAddTestCache) Destroy() {}
func (e *GetAddTestCache) Size() int {
	return 0
}

func (e *GetAddTestCache) GC() {}
func (e *GetAddTestCache) Snapshot() []*scheduler.TaskSnapshot {
	return nil
}

func (e *GetAddTestCache) Evict(resources *Resources) bool {
	return false
}
