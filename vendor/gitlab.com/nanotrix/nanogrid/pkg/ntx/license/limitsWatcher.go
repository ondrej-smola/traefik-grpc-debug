package license

import (
	"sync"
)

type LimitsWatcher interface {
	Limits() *Limits
}

type LimitsWatcherFn func() *Limits

func (f LimitsWatcherFn) Limits() *Limits {
	return f()
}

type DefaultLimitsWatcher struct {
	lock *sync.Mutex
	l    *Limits
}

var _ = LimitsWatcher(&DefaultLimitsWatcher{})

func NewDefaultLimitsWatcher(l *Limits) *DefaultLimitsWatcher {
	return &DefaultLimitsWatcher{
		lock: &sync.Mutex{},
		l:    l.Clone(),
	}
}

func (l *DefaultLimitsWatcher) Synchronize(w Watcher) error {
	return w.Watch(func(lic *License) {
		l.lock.Lock()
		l.l = lic.Limits.Clone()
		l.lock.Unlock()
	})
}

func (l *DefaultLimitsWatcher) Limits() *Limits {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.l.Clone()
}
