package license

import (
	"context"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/zk"
)

type Watcher interface {
	Watch(cb WatchCallback) error
}

type WatcherFn func(cb WatchCallback) error

func (f WatcherFn) Watch(cb WatchCallback) error {
	return f(cb)
}

type WatchCallback func(l *License)

func NewZkWatcher(store Store, w zk.Client, ctx context.Context) Watcher {
	return WatcherFn(func(cb WatchCallback) error {
		lic, err := store.License()
		if err != nil {
			return err
		}

		cb(lic)

		cbErr := make(chan error, 1)
		runErr := make(chan error, 1)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		go func() {
			runErr <- w.Watch(ZkActiveLicensePath, func(kv *zk.KVPair) {
				if lic, err := store.License(); err != nil {
					select {
					case cbErr <- err:
					default:
					}
				} else {
					cb(lic)
				}
			}, ctx)
		}()

		select {
		case err := <-runErr:
			return err
		case err := <-cbErr:
			return err
		}
	})
}
