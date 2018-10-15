package shared

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler/cache"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
)

type (
	Keeper interface {
		Get(key string) (t scheduler.Task, found bool)
		Ping(ids ...string)
		Release(ids ...string)
		Kill(ids ...string)
		GC()
		Snapshot() []*scheduler.TaskSnapshot
	}

	keeperEntry struct {
		task     scheduler.Task
		cacheKey string
		lastPing time.Time
	}

	DefaultKeeper struct {
		cache   cache.TaskCache
		gcAfter time.Duration

		log log.Logger
		sync.Mutex
		//mutex for
		borrowed []*keeperEntry
	}

	DefaultKeeperOpt func(d *DefaultKeeper)
)

func WithKeeperLogger(log log.Logger) DefaultKeeperOpt {
	return func(d *DefaultKeeper) {
		d.log = log
	}
}

func WithKeeperGCAfter(after time.Duration) DefaultKeeperOpt {
	return func(d *DefaultKeeper) {
		d.gcAfter = after
	}
}

func NewKeeper(cache cache.TaskCache, opts ...DefaultKeeperOpt) *DefaultKeeper {
	k := &DefaultKeeper{
		cache:   cache,
		gcAfter: 10 * time.Second,
		log:     log.NewNopLogger(),
	}

	for _, o := range opts {
		o(k)
	}

	return k
}

var _ = Keeper(&DefaultKeeper{})

func (k *DefaultKeeper) Get(key string) (scheduler.Task, bool) {
	t, ok := k.cache.Get(key)

	if ok {
		k.Lock()
		k.borrowed = append(k.borrowed, &keeperEntry{
			task:     t,
			cacheKey: key,
			lastPing: timeutil.NowUTC(),
		})
		k.Unlock()
	}

	return t, ok
}

func (k *DefaultKeeper) GC() {
	k.Lock()
	defer k.Unlock()

	now := timeutil.NowUTC().Add(-k.gcAfter)

	var cpy []*keeperEntry
	for _, b := range k.borrowed {
		if b.lastPing.Before(now) {
			b.task.Kill()
			k.log.Log("ev", "kill_task", "tid", b.task.Snapshot().Id, "cause", "gc", "expired", b.lastPing)
		} else {
			cpy = append(cpy, b)
		}
	}

	k.borrowed = cpy
}

func idsToMap(ids []string) map[string]bool {
	mp := make(map[string]bool)

	for _, id := range ids {
		mp[id] = true
	}

	return mp
}

func (k *DefaultKeeper) Ping(ids ...string) {
	mp := idsToMap(ids)

	k.Lock()
	defer k.Unlock()

	now := timeutil.NowUTC()

	for _, b := range k.borrowed {
		if mp[b.task.Snapshot().Id] {
			b.lastPing = now
		}
	}
}

func (k *DefaultKeeper) Release(ids ...string) {
	mp := idsToMap(ids)

	k.Lock()
	defer k.Unlock()

	var filter []*keeperEntry

	for _, b := range k.borrowed {
		snap := b.task.Snapshot()

		if mp[snap.Id] {
			if snap.State.IsRunning() {
				k.log.Log("ev", "returned_task_to_cache", "tid", snap.Id)
				k.cache.Add(b.cacheKey, b.task)
			} else {
				b.task.Release()
				k.log.Log("ev", "release_task", "tid", snap.Id, "cause", "not_running", "state", snap.State)
			}
		} else {
			filter = append(filter, b)
		}
	}

	k.borrowed = filter
}

func (k *DefaultKeeper) Kill(ids ...string) {
	mp := idsToMap(ids)

	k.Lock()
	defer k.Unlock()

	var filter []*keeperEntry

	for _, b := range k.borrowed {
		snap := b.task.Snapshot()
		if mp[snap.Id] {
			b.task.Kill()
			k.log.Log("ev", "kill_task", "tid", snap.Id)
		} else {
			filter = append(filter, b)
		}
	}

	k.borrowed = filter
}

func (k *DefaultKeeper) Snapshot() []*scheduler.TaskSnapshot {
	k.Lock()
	defer k.Unlock()

	res := make([]*scheduler.TaskSnapshot, len(k.borrowed))

	for i, b := range k.borrowed {
		res[i] = b.task.Snapshot()
	}

	return res
}
