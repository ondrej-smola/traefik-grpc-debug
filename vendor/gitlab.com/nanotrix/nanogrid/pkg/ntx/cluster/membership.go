package cluster

import (
	"encoding/json"
	"sync"

	"context"
	"path"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/zk"
)

type (
	Listing interface {
		List() Members
	}

	ListingFn func() Members

	Membership interface {
		// add member to member list and block till connection to cluster lost
		Join(m Member, ctx context.Context) error
		Watch(ctx context.Context) error
		Listing
	}

	ZkDetector struct {
		cl  zk.Client
		dir string
		log log.Logger

		mtx     *sync.Mutex
		members Members
	}

	Static struct {
		members Members
		memLock *sync.Mutex
	}
)

var _ = Listing(&ZkDetector{})
var _ = Membership(&ZkDetector{})

func (f ListingFn) List() Members {
	return f()
}

func NewEmpty() *Static {
	return NewStatic(nil)
}

func NewStatic(m Members) *Static {
	return &Static{members: m, memLock: &sync.Mutex{}}
}

func (r *Static) List() Members {
	r.memLock.Lock()
	defer r.memLock.Unlock()

	return r.members.Clone()
}

func (r *Static) Watch(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (r *Static) Join(m Member, ctx context.Context) error {
	if err := m.Valid(); err != nil {
		return errors.Wrap(err, "member")
	}

	r.memLock.Lock()
	r.members = append(r.members, m)
	r.memLock.Unlock()

	return r.Watch(ctx)
}

func NewZkDetector(zk zk.Client, dir string) *ZkDetector {
	return &ZkDetector{
		cl:  zk,
		dir: dir,
		mtx: &sync.Mutex{},
	}
}

func (z *ZkDetector) List() Members {
	z.mtx.Lock()
	defer z.mtx.Unlock()

	cpy := make([]Member, len(z.members))
	copy(cpy, z.members)
	return cpy
}

func (z *ZkDetector) kvsToMembers(kvs []*zk.KVPair) (Members, error) {
	members := make([]Member, len(kvs))

	for i, kv := range kvs {
		member := Member{}

		if err := json.Unmarshal(kv.Value, &member); err != nil {
			return nil, errors.Wrapf(err, "json[%v]: key[%v]", i, kv.Key)
		} else if member.Valid(); err != nil {
			return nil, errors.Wrapf(err, "member[%v]: key[%v]", i, kv.Key)
		} else {
			members[i] = member
		}
	}

	return members, nil
}

func (z *ZkDetector) connect(ctx context.Context) error {
	cbErr := make(chan error, 1)
	runErr := make(chan error, 1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		runErr <- z.cl.WatchTree(z.dir, func(kvs zk.KVPairs) {
			newMembers, err := z.kvsToMembers(kvs)
			if err != nil {
				select {
				case cbErr <- err:
				default:
				}
			}
			z.mtx.Lock()
			z.members = newMembers
			z.mtx.Unlock()
		}, ctx)
	}()

	select {
	case err := <-runErr:
		return err
	case err := <-cbErr:
		return err
	}
}

func (z *ZkDetector) Watch(ctx context.Context) error {
	return z.connect(ctx)
}

func (z *ZkDetector) Join(m Member, ctx context.Context) error {
	if err := m.Valid(); err != nil {
		return errors.Wrap(err, "member")
	}

	mBody, err := json.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "json marshall")
	}

	if err := z.cl.AtomicPut(path.Join(z.dir, m.Id), mBody, nil, zk.ZkFlagEphemeral); err != nil {
		return errors.Wrap(err, "add self to member list")
	}

	return z.connect(ctx)
}

func Filter(in Listing, accept func(m Member) bool) Listing {
	return ListingFn(func() Members {
		members := in.List()

		res := Members{}
		for _, m := range members {
			if accept(m) {
				res = append(res, m)
			}
		}

		return res
	})
}

func RemoveSelfFromListing(selfId string, in Listing) Listing {
	return ListingFn(func() Members {
		members := in.List()

		res := Members{}
		for _, m := range members {
			if m.Id != selfId {
				res = append(res, m)
			}
		}

		return res
	})
}
