package cache

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
)

type (
	Line interface {
		Add(scheduler.Task)
		Get() (t scheduler.Task, found bool)
		Snapshot() []*scheduler.TaskSnapshot
		Size() int
		GC()
		Destroy()
	}

	expiringLineEntry struct {
		created time.Time
		expire  time.Time
		val     scheduler.Task
	}

	LRULine struct {
		ttl     time.Duration
		maxSize int

		cache []*expiringLineEntry
		log   log.Logger

		sync.Mutex
	}

	LRULineOpt func(t *LRULine)
)

func WithLRULineMaxSize(size int) LRULineOpt {
	return func(t *LRULine) {
		t.maxSize = size
	}
}

func WithLRULineTTL(ttl time.Duration) LRULineOpt {
	if ttl < 1*time.Millisecond {
		panic(fmt.Sprintf("TTL must be >= 1 millisecond, got %v", ttl))
	}

	return func(t *LRULine) {
		t.ttl = ttl
	}
}

func WithLRULineLogger(log log.Logger) LRULineOpt {
	return func(t *LRULine) {
		t.log = log
	}
}

func NewLRULine(opts ...LRULineOpt) *LRULine {
	r := &LRULine{
		ttl:     time.Minute,
		maxSize: math.MaxInt64,
		log:     log.NewNopLogger(),
	}

	for _, o := range opts {
		o(r)
	}

	return r
}

func (t *LRULine) GC() {
	now := timeutil.NowUTC()
	// filter without allocating new array
	t.Lock()
	res := t.cache[:0] // reuse underlying array
	for _, v := range t.cache {

		snap := v.val.Snapshot()
		release := false

		if !snap.State.IsRunning() {
			t.log.Log("ev", "release_local_cached_task", "tid", snap.Id, "cause", "not_running", "state", snap.State)
			release = true
		} else if v.expire.Before(now) {
			t.log.Log("ev", "release_expired_local_cached_task", "tid", snap.Id, "cause", "expired", "at", v.expire)
			release = true
		}

		if release {
			v.val.Release()
		} else {
			res = append(res, v)
		}
	}
	t.cache = res
	t.Unlock()
}

func (t *LRULine) Add(val scheduler.Task) {
	t.Lock()
	if len(t.cache) < t.maxSize {
		now := timeutil.NowUTC()
		expire := now.Add(t.ttl)
		t.log.Log("ev", "cached_task", "tid", val.Snapshot().Id, "expire", expire)
		t.cache = append(t.cache, &expiringLineEntry{expire: expire, created: now, val: val})
	} else {
		t.log.Log("ev", "release_task", "tid", val.Snapshot().Id, "cause", "cache_full")
		val.Release()
	}
	t.Unlock()
}

func (t *LRULine) Get() (scheduler.Task, bool) {
	t.Lock()
	r, f := t.getUnsafe()
	t.Unlock()
	return r, f
}

func (t *LRULine) getUnsafe() (scheduler.Task, bool) {
	for i := range t.cache {
		v := t.cache[len(t.cache)-1-i]
		if v.val.Snapshot().State.IsRunning() {
			t.cache = t.cache[:len(t.cache)-1-i]
			return v.val, true
		}
	}

	t.cache = nil
	return nil, false
}

func (t *LRULine) Destroy() {
	t.Lock()
	for _, v := range t.cache {
		t.log.Log("ev", "release_local_cached_task", "tid", v.val.Snapshot().Id, "cause", "line_destroy")
		v.val.Release()
	}
	t.cache = nil
	t.Unlock()
}

func (t *LRULine) Size() int {
	t.Lock()
	size := len(t.cache)
	t.Unlock()
	return size
}

func (t *LRULine) Snapshot() []*scheduler.TaskSnapshot {
	t.Lock()
	ret := make([]*scheduler.TaskSnapshot, len(t.cache))
	for i, c := range t.cache {
		ret[i] = c.val.Snapshot()
		ret[i].Created = uint64(c.created.Unix())
		ret[i].Expire = uint64(c.expire.Unix())
	}

	t.Unlock()
	return ret
}
