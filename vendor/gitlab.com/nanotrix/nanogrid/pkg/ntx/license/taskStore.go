package license

import (
	"fmt"
	"sync"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth/perm"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

const (
	versionNotSet = "version_not_set"
)

type (
	TaskStore interface {
		TaskProfileSearch
		Get(id string) (*Task_Profile, error)
		Snapshot() (*VersionedTasks, error)
	}

	TaskProfileSearch interface {
		Find(label string, tags ...string) (*Task_Profile, error)
	}

	TaskProfileSearchFn func(label string, tags ...string) (*Task_Profile, error)

	InMemTaskStore struct {
		lock *sync.RWMutex

		expire  uint64
		index   map[string]*Task_Profile
		content *VersionedTasks
	}
)

func (f TaskProfileSearchFn) Find(label string, tags ...string) (*Task_Profile, error) {
	return f(label, tags...)
}

func NewInMemTaskStore(content *VersionedTasks) *InMemTaskStore {
	s := &InMemTaskStore{
		lock:    &sync.RWMutex{},
		index:   make(map[string]*Task_Profile),
		content: &VersionedTasks{Version: versionNotSet, Created: timeutil.NowUTCUnixSeconds()},
	}

	s.set(content)
	return s
}

var _ = TaskStore(&InMemTaskStore{})

func (s *InMemTaskStore) Snapshot() (*VersionedTasks, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.content.Clone(), nil
}

func (s *InMemTaskStore) Find(label string, tags ...string) (*Task_Profile, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, i := range s.content.Tasks {
		if i.ContainsAllTags(tags...) {
			for _, p := range i.Profiles {
				if p.ContainsLabel(label) {
					return p.Clone(), nil
				}
			}
		}
	}

	return nil, util.NewNotFoundError(fmt.Sprintf("Entry not found for tags: %v", tags))
}

func (s *InMemTaskStore) Get(id string) (*Task_Profile, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if p, ok := s.index[id]; !ok {
		return nil, util.NewNotFoundError(fmt.Sprintf("Task %v not found", id))
	} else {
		return p.Clone(), nil
	}
}

func (s *InMemTaskStore) set(t *VersionedTasks) {
	tasks := t.Clone()

	s.lock.Lock()
	defer s.lock.Unlock()

	s.content = tasks
	s.index = make(map[string]*Task_Profile)
	for _, it := range s.content.Tasks {
		for _, p := range it.Profiles {
			s.index[p.Id] = p
		}
	}
}

func (s *InMemTaskStore) Synchronize(watcher Watcher) error {
	return watcher.Watch(func(l *License) {
		s.set(l.Tasks)
	})
}

func (t *Task) ContainsAllTags(tags ...string) bool {
	for _, it := range tags {
		found := false
		for _, t2 := range t.Tags {
			if it == t2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (t *Task_Profile) ContainsLabel(label string) bool {
	for _, l := range t.Labels {
		if l == label {
			return true
		}
	}

	return false
}

func RemoveTaskConfigurations(tasks []*Task) {
	for _, t := range tasks {
		for _, p := range t.Profiles {
			p.Task = nil
		}
	}
}

func FilterTasks(tasks []*Task, perm perm.Permissions) []*Task {
	var res []*Task

	for _, t := range tasks {
		tCopy := t.Clone()
		tCopy.Profiles = nil

		for _, p := range t.Profiles {
			if perm.RunTask(p.Id) {
				tCopy.Profiles = append(tCopy.Profiles, p)
			}
		}

		if len(tCopy.Profiles) > 0 {
			res = append(res, tCopy)
		}
	}

	return res
}
