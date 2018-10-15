package license

import (
	"sync/atomic"
)

type Expiration interface {
	// unix seconds
	Expire() uint64
}

type ExpirationWatcherFn func() uint64

func (f ExpirationWatcherFn) Expire() uint64 {
	return f()
}

type ExpirationFromWatcher struct {
	expire uint64
}

var _ = Expiration(&ExpirationFromWatcher{})

func NewDefaultExpirationWatcher(expire uint64) *ExpirationFromWatcher {
	return &ExpirationFromWatcher{expire: expire}
}

func (l *ExpirationFromWatcher) Synchronize(w Watcher) error {
	return w.Watch(func(lic *License) {
		atomic.StoreUint64(&l.expire, lic.Expire)
	})
}

func (l *ExpirationFromWatcher) Expire() uint64 {
	return atomic.LoadUint64(&l.expire)
}
