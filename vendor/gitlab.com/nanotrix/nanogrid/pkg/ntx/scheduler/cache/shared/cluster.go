package shared

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/grpcutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
)

type (
	Cluster interface {
		CacheQuery
		SetMembers(members []string, ctx context.Context)
		HealthCheck(ctx context.Context)
		Snapshot() []*scheduler.TaskSnapshot
	}

	clusterMember struct {
		cache Cache

		failedPings int
	}

	cluster struct {
		prov CacheProvider

		members                 map[string]*clusterMember
		removeAfterNFailedPings int
		log                     log.Logger

		sync.RWMutex
	}

	noopCluster struct {
	}

	CoordinatorOpt func(c *cluster)
)

func WithClusterRemoveAfterNFailedPings(n int) CoordinatorOpt {
	if n <= 0 {
		panic(fmt.Sprintf("n must be > 0, got %v", n))
	}

	return func(c *cluster) {
		c.removeAfterNFailedPings = n

	}
}

func WithClusterLogger(log log.Logger) CoordinatorOpt {
	return func(c *cluster) {
		c.log = log
	}
}

func NewNoopCluster() Cluster {
	return &noopCluster{}
}

func NewCluster(prov CacheProvider, opts ...CoordinatorOpt) Cluster {
	c := &cluster{
		prov:                    prov,
		members:                 map[string]*clusterMember{},
		removeAfterNFailedPings: 5,
		log:                     log.NewNopLogger(),
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *cluster) SetMembers(members []string, ctx context.Context) {
	c.Lock()

	updatedMembers := map[string]bool{}

	for _, m := range members {
		updatedMembers[m] = true

		_, ok := c.members[m]
		if ok {
			continue
		}

		newCache, err := c.prov(m)
		if err != nil {
			c.log.Log("ev", "failed_to_create_cache", "endpoint", m, "err", err)
			continue
		}
		c.members[m] = &clusterMember{
			cache:       newCache,
			failedPings: 0,
		}

		c.log.Log("ev", "new_member", "endpoint", m)
	}

	toDelete := map[string]*clusterMember{}

	for m, v := range c.members {
		if _, ok := updatedMembers[m]; !ok {
			delete(c.members, m)
			toDelete[m] = v.clone()
		}
	}

	c.Unlock()

	for m, v := range toDelete {
		if err := v.cache.Destroy(ctx); err != nil {
			c.log.Log("ev", "failed_to_remove_member", "id", m, "err", err)
		} else {
			c.log.Log("ev", "member_removed", "id", m)
		}
	}
}

func (c *cluster) HealthCheck(ctx context.Context) {
	members := c.membersCopy()

	wg := &sync.WaitGroup{}
	wg.Add(len(members))

	for k, v := range members {
		go func(endpoint string, ce *clusterMember) {
			c.checkMember(endpoint, ce, ctx)
			wg.Done()
		}(k, v)
	}

	wg.Wait()
}

func (c *cluster) checkMember(endpoint string, ce *clusterMember, ctx context.Context) {
	err := ce.cache.Ping(ctx)

	failedPings := ce.failedPings

	if err != nil {
		failedPings += 1
		c.log.Log("ev", "ping_failed", "endpoint", endpoint, "err", err)
	} else {
		if failedPings > 0 {
			c.log.Log("ev", "healthy_again", "endpoint", endpoint)
		}
		failedPings = 0
	}

	if failedPings >= c.removeAfterNFailedPings {
		c.Lock()
		delete(c.members, endpoint)
		c.Unlock()

		if err := errors.Wrap(ce.cache.Destroy(ctx), "destroy taskCache"); err != nil {
			c.log.Log("ev", "failed_to_remove_member", "endpoint", endpoint, "err", err)
		} else {
			c.log.Log("ev", "member_removed", "endpoint", endpoint)
		}
	} else {
		c.Lock()
		el, ok := c.members[endpoint]
		if ok {
			el.failedPings = failedPings
		}
		c.Unlock()
	}
}

func (c *cluster) Get(req *TaskRequest, ctx context.Context) (scheduler.Task, bool, error) {
	members := c.membersCopy()

	for k, v := range members {
		if v.available() {
			t, ok, err := v.cache.Get(req, ctx)
			if err != nil {
				if !grpcutil.IsContextCancelled(err) {
					c.log.Log("ev", "remote_cache_get_err", "endpoint", k, "err", err)
				}
			} else if ok {
				c.log.Log("ev", "remote_cache_hit", "endpoint", k, "tid", t.Snapshot().Id, "debug", true)
				return t, true, nil
			}
		}
	}

	return nil, false, nil
}

func (c *cluster) membersCopy() map[string]*clusterMember {
	c.RLock()
	defer c.RUnlock()
	cpy := map[string]*clusterMember{}

	for k, v := range c.members {
		cpy[k] = v.clone()
	}

	return cpy
}

func (c *clusterMember) available() bool {
	return c.failedPings == 0
}

func (c *clusterMember) clone() *clusterMember {
	return &clusterMember{
		cache:       c.cache,
		failedPings: c.failedPings,
	}
}

func (c *cluster) Snapshot() []*scheduler.TaskSnapshot {
	c.Lock()
	defer c.Unlock()

	res := []*scheduler.TaskSnapshot{}
	for _, m := range c.members {
		res = append(res, m.cache.Snapshot()...)
	}

	return res
}

func (c *noopCluster) Snapshot() []*scheduler.TaskSnapshot { return nil }
func (c *noopCluster) Get(req *TaskRequest, ctx context.Context) (scheduler.Task, bool, error) {
	return nil, false, nil
}
func (c *noopCluster) SetMembers(members []string, ctx context.Context) {}
func (c *noopCluster) HealthCheck(ctx context.Context)                  {}
