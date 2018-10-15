package zk

import (
	"context"
	"math/rand"
	"path"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/samuel/go-zookeeper/zk"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

// for testing purposes

type keyWatcherEntry struct {
	cb  WatchKeyCallback
	ctx context.Context
	err chan error

	sync.Mutex
}

type dirWatcherEntry struct {
	cb  WatchTreeCallback
	ctx context.Context
	err chan error

	sync.Mutex
}

type InMemZk struct {
	lock *sync.Mutex
	// path -> *KVPair
	content map[string]*KVPair

	keyWatch map[string][]*keyWatcherEntry
	dirWatch map[string][]*dirWatcherEntry
}

var _ = Client(&InMemZk{})

func NewInMemZk() *InMemZk {
	return &InMemZk{
		lock:     &sync.Mutex{},
		content:  make(map[string]*KVPair),
		keyWatch: make(map[string][]*keyWatcherEntry),
		dirWatch: make(map[string][]*dirWatcherEntry),
	}
}

func (z *InMemZk) Get(key string) (pair *KVPair, err error) {
	z.lock.Lock()
	defer z.lock.Unlock()

	if act, ok := z.content[key]; !ok {
		return nil, ErrKeyNotFound
	} else {
		return act.Clone(), nil
	}
}

func (z *InMemZk) AtomicPut(key string, value []byte, previous *KVPair, flags ...Flag) error {
	z.lock.Lock()
	defer z.lock.Unlock()
	return z.atomicPut(key, value, previous, flags...)
}

func (z *InMemZk) Put(key string, value []byte) error {
	z.lock.Lock()
	defer z.lock.Unlock()
	return z.atomicPut(key, value, z.content[key])
}

func (z *InMemZk) atomicPut(key string, value []byte, previous *KVPair, flags ...Flag) error {
	act, ok := z.content[key]

	var updated *KVPair

	if ok {
		if act.Stat.Version != previous.Stat.Version {
			return ErrKeyModified
		}

		updated = &KVPair{
			Key:   key,
			Value: value,
			Stat:  act.Clone().Stat,
		}

		z.setModified(updated, value)
	} else {
		if previous != nil {
			return ErrKeyModified
		}

		updated = z.newKvPair(key, value)
	}

	z.content[key] = updated.Clone()
	z.notify(key, updated)

	return nil
}

func (z *InMemZk) newKvPair(k string, v []byte) *KVPair {
	kv := &KVPair{Key: k, Value: v, Stat: &zk.Stat{}}
	kv.Stat.Ctime = int64(timeutil.NowUTCUnixSeconds() * 1000)
	kv.Stat.Mtime = kv.Stat.Ctime
	kv.Stat.Czxid = rand.Int63()
	kv.Stat.Mzxid = kv.Stat.Czxid
	kv.Stat.DataLength = int32(len(v))
	return kv
}

func (z *InMemZk) setModified(kv *KVPair, v []byte) {
	kv.Value = util.CloneBytes(v)
	kv.Stat.Version += 1
	kv.Stat.Mtime = int64(timeutil.NowUTCUnixSeconds() * 1000)
	kv.Stat.Mzxid = rand.Int63()
	kv.Stat.DataLength = int32(len(v))
}

func (z *InMemZk) Delete(key string) error {
	z.lock.Lock()
	defer z.lock.Unlock()

	if _, ok := z.content[key]; !ok {
		return ErrKeyNotFound
	} else {
		delete(z.content, key)
	}

	for _, e := range z.keyWatch[key] {
		e.err <- errors.New("key deleted")
	}

	delete(z.keyWatch, key)

	dir := z.getDir(key)
	for _, e := range z.dirWatch[dir] {
		e.err <- errors.New("dir deleted")
	}

	delete(z.dirWatch, key)

	return nil
}

func (z *InMemZk) List(key string) ([]*KVPair, error) {
	z.lock.Lock()
	defer z.lock.Unlock()

	pathS := strings.TrimSuffix(key, "/") + "/?*"

	if z.content[key] == nil {
		return nil, ErrKeyNotFound
	}

	var res []*KVPair

	for k, v := range z.content {
		if ok, err := path.Match(pathS, k); err != nil {
			return nil, errors.Wrap(err, "match path")
		} else if ok {
			res = append(res, v.Clone())
		}
	}

	return res, nil
}

func (z *InMemZk) notify(key string, p *KVPair) {
	z.notifyKeyWatchers(key, p)
	z.notifyDirWatchers(z.getDir(key))
}

func (z *InMemZk) getDir(key string) string {
	parts := SplitPath(key)
	if len(parts) > 1 {
		return Join(parts[:len(parts)-1]...)
	} else {
		return key
	}
}

func (z *InMemZk) notifyKeyWatchers(key string, p *KVPair) {
	var filtered []*keyWatcherEntry

	for _, e := range z.keyWatch[key] {
		cpy := e
		if cpy.ctx.Err() == nil {
			go func() {
				cpy.Lock()
				cpy.cb(p.Clone())
				cpy.Lock()
			}()
			filtered = append(filtered, cpy)
		}
	}

	z.keyWatch[key] = filtered
}

func (z *InMemZk) listChildren(dir string) KVPairs {
	var kvs KVPairs

	match := strings.TrimSuffix(dir, "/") + "/?*"

	for k, v := range z.content {
		if ok, _ := path.Match(match, k); ok {
			kvs = append(kvs, v)
		}
	}
	return kvs
}

func (z *InMemZk) notifyDirWatchers(dir string) {
	kvs := z.listChildren(dir)

	var filtered []*dirWatcherEntry

	for _, e := range z.dirWatch[dir] {
		cpy := e
		if cpy.ctx.Err() == nil {
			go func() {
				cpy.Lock()
				cpy.cb(kvs.Clone())
				cpy.Lock()
			}()
			filtered = append(filtered, cpy)
		}
	}

	z.dirWatch[dir] = filtered
}

func (z *InMemZk) Watch(key string, cb WatchKeyCallback, ctx context.Context) error {
	z.lock.Lock()

	kv := z.content[key]
	if kv == nil {
		z.lock.Unlock()
		return ErrKeyNotFound
	}

	cpy := kv.Clone()
	errChan := make(chan error, 1)
	entry := &keyWatcherEntry{cb: cb, ctx: ctx, err: errChan}

	z.keyWatch[key] = append(z.keyWatch[key], entry)
	entry.Lock()
	z.lock.Unlock()
	entry.cb(cpy)
	entry.Unlock()

	select {
	case <-ctx.Done():
		z.lock.Lock()
		delete(z.keyWatch, key)
		z.lock.Unlock()
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func (z *InMemZk) WatchTree(key string, cb WatchTreeCallback, ctx context.Context) error {
	z.lock.Lock()

	if z.content[key] == nil {
		return ErrKeyNotFound
	}

	child := z.listChildren(key).Clone()
	errChan := make(chan error, 1)
	entry := &dirWatcherEntry{cb: cb, ctx: ctx, err: errChan}

	z.dirWatch[key] = append(z.dirWatch[key], entry)
	entry.Lock()
	z.lock.Unlock()
	entry.cb(child)
	entry.Unlock()

	select {
	case <-ctx.Done():
		z.lock.Lock()
		delete(z.keyWatch, key)
		z.lock.Unlock()
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}
