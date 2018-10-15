package zk

import (
	"strings"
	"time"

	"fmt"

	"context"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/samuel/go-zookeeper/zk"
)

type Flag int

const (
	ZkFlagEphemeral = Flag(zk.FlagEphemeral)
	FlagSequence    = Flag(zk.FlagSequence)
)

func mergeFlags(f ...Flag) int32 {
	res := int32(0)

	for _, fi := range f {
		res |= int32(fi)
	}

	return res
}

type WatchKeyCallback func(k *KVPair)
type WatchTreeCallback func(k KVPairs)

type Client interface {
	Get(key string) (pair *KVPair, err error)
	Put(key string, value []byte) error
	AtomicPut(key string, value []byte, prev *KVPair, zFlags ...Flag) error
	Delete(key string) error
	List(key string) ([]*KVPair, error)
	Watch(key string, cb WatchKeyCallback, ctx context.Context) error
	WatchTree(key string, cb WatchTreeCallback, ctx context.Context) error
}

// Zookeeper is the receiver type for
// the Store interface
type ClientImpl struct {
	timeout time.Duration
	client  *zk.Conn

	log log.Logger
}

var _ = Client(&ClientImpl{})

type logToGoKitLog struct {
	log log.Logger
}

func (l *logToGoKitLog) Printf(format string, a ...interface{}) {
	l.log.Log("ev", fmt.Sprintf(format, a...))
}

type ZkOpt func(z *ClientImpl)

func WithLogger(log log.Logger) ZkOpt {
	return func(z *ClientImpl) {
		z.log = log
	}
}

func WithSessionTimeout(timeout time.Duration) ZkOpt {
	return func(z *ClientImpl) {
		z.timeout = timeout
	}
}

// SplitPath splits the key to extract path informations
func SplitPath(key string) (path []string) {
	if strings.Contains(key, "/") {
		path = strings.Split(key, "/")
		// if key is absolute first part will be blank string - replace it with '/'
		if path[0] == "" {
			path[0] = "/"
		}
	} else {
		path = []string{key}
	}
	return path
}

// join the path parts with '/'
func Join(parts ...string) string {
	res := strings.Join(parts, "/")
	// handle case Join("/","a")
	if strings.HasPrefix(res, "//") {
		return res[1:]
	} else {
		return res
	}
}

var (
	// ErrKeyModified is thrown during an atomic operation if the index does not match the one in the store
	ErrKeyModified = errors.New("unable to complete atomic operation, key modified")
	// ErrKeyNotFound is thrown when the key is not found in the store during a Get operation
	ErrKeyNotFound = errors.New("key not found")
	// ErrKeyExists is thrown when the previous value exists in the case of an AtomicPut
	ErrKeyExists = errors.New("previous K/V pair exists, cannot complete atomic operation")
)

// New creates a new Zookeeper ClientImpl given a
// list of endpoints
func New(endpoints []string, opts ...ZkOpt) (*ClientImpl, error) {
	s := &ClientImpl{
		timeout: 10 * time.Second,
		log:     log.NewNopLogger(),
	}

	for _, o := range opts {
		o(s)
	}
	// WatchCluster to Zookeeper
	conn, _, err := zk.Connect(endpoints, s.timeout)
	if err != nil {
		return nil, err
	}
	s.client = conn
	s.client.SetLogger(&logToGoKitLog{
		log: s.log,
	})

	return s, nil
}

// setTimeout sets the TestClientTimeout for connecting to Zookeeper
func (s *ClientImpl) setTimeout(time time.Duration) {
	s.timeout = time
}

// Get the value at "key", returns the last modified index
// to use in conjunction to Atomic calls
func (s *ClientImpl) Get(key string) (pair *KVPair, err error) {
	resp, stat, err := s.client.Get(key)

	if err != nil {
		if err == zk.ErrNoNode {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	pair = &KVPair{
		Key:   key,
		Value: resp,
		Stat:  stat,
	}

	return pair, nil
}

// createFullPath creates the entire path for a directory
// that does not exist
func (s *ClientImpl) createFullPath(path []string, ephemeral bool) error {
	for i := 1; i <= len(path); i++ {
		newpath := Join(path[:i]...)
		if i == len(path) && ephemeral {
			_, err := s.client.Create(newpath, []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
			return err
		}
		_, err := s.client.Create(newpath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			// Skip if node already exists
			if err != zk.ErrNodeExists {
				return err
			}
		}
	}
	return nil
}

func (s *ClientImpl) Put(key string, value []byte) error {
	_, err := s.client.Set(key, value, -1)

	if err == zk.ErrNoNode {
		parts := SplitPath(strings.TrimSuffix(key, "/"))
		parts = parts[:len(parts)-1]
		if err := s.createFullPath(parts, false); err != nil {
			return errors.Wrap(err, "created dir path")
		}

		_, err = s.client.Set(key, value, -1)
	}

	return err
}

// Delete a value at "key"
func (s *ClientImpl) Delete(key string) error {
	err := s.client.Delete(key, -1)
	if err == zk.ErrNoNode {
		return ErrKeyNotFound
	}
	return err
}

// Exists checks if the key exists inside the store
func (s *ClientImpl) Exists(key string) (bool, error) {
	exists, _, err := s.client.Exists(key)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// WatchMembers for changes on a "key"
// It returns a channel that will receive changes or pass
// on errors. Upon creation, the current value will first
// be sent to the channel. Providing a non-nil stopCh can
// be used to stop watching.
func (s *ClientImpl) Watch(key string, cb WatchKeyCallback, ctx context.Context) error {
	for {
		body, stat, eventCh, err := s.client.GetW(key)
		if err != nil {
			return err
		}

		cb(&KVPair{
			Key:   key,
			Value: body,
			Stat:  stat,
		})

		select {
		case <-eventCh:
		case <-ctx.Done():
			// There is no way to stop GetW so just quit
			return ctx.Err()
		}
	}
}

// WatchTree watches for changes on a "directory"
// It returns a channel that will receive changes or pass
// on errors. Upon creating a watch, the current childs values
// will be sent to the channel .Providing a non-nil stopCh can
// be used to stop watching.
func (s *ClientImpl) WatchTree(directory string, cb WatchTreeCallback, ctx context.Context) error {
	for {
		keys, _, eventCh, err := s.client.ChildrenW(directory)
		if err != nil {
			return errors.Wrapf(err, "watch dir: %v", directory)
		}

		kvs, err := s.listKeysInDir(directory, keys...)
		if err != nil {
			s.log.Log("ev", "keys_in_dir", "dir", directory, "err", err)
			return errors.Wrapf(err, "keys in dir: %v", directory)
		}

		cb(kvs)

		select {
		case ev := <-eventCh:
			if ev.State == zk.StateDisconnected {
				return errors.New("watcher disconnected")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *ClientImpl) listKeysInDir(dir string, keys ...string) ([]*KVPair, error) {
	var kvs []*KVPair

	for _, key := range keys {
		pair, err := s.Get(Join(strings.TrimSuffix(dir, "/"), key))
		if err != nil {
			// If node is not found: List is out of date, retry
			if err == ErrKeyNotFound {
				continue
			} else {
				return nil, err
			}
		}

		kvs = append(kvs, &KVPair{
			Key:   key,
			Value: []byte(pair.Value),
			Stat:  pair.Stat,
		})
	}

	return kvs, nil
}

// List child nodes of a given directory
func (s *ClientImpl) List(directory string) ([]*KVPair, error) {
	keys, _, err := s.client.Children(directory)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	return s.listKeysInDir(directory, keys...)
}

// DeleteTree deletes a range of keys under a given directory
func (s *ClientImpl) DeleteTree(directory string) error {
	pairs, err := s.List(directory)
	if err != nil {
		return err
	}

	var reqs []interface{}

	for _, pair := range pairs {
		reqs = append(reqs, &zk.DeleteRequest{
			Path:    directory + "/" + pair.Key,
			Version: -1,
		})
	}

	_, err = s.client.Multi(reqs...)
	return err
}

// AtomicPut put a value at "key" if the key has not been
// modified in the meantime, throws an error if this is the case
func (s *ClientImpl) AtomicPut(key string, value []byte, prev *KVPair, zkFlags ...Flag) error {
	flags := mergeFlags(zkFlags...)

	if prev != nil {
		_, err := s.client.Set(key, value, prev.Stat.Version)
		if err != nil {
			if err == zk.ErrNoNode {
				return ErrKeyNotFound
			} else if err == zk.ErrBadVersion {
				return ErrKeyModified
			} else {
				return err
			}
		}
	} else {
		// Interpret previous == nil as create operation.
		_, err := s.client.Create(key, value, flags, zk.WorldACL(zk.PermAll))
		if err != nil {
			// Directory does not exist
			if err == zk.ErrNoNode {

				parts := SplitPath(strings.TrimSuffix(key, "/"))
				parts = parts[:len(parts)-1]
				if err := s.createFullPath(parts, false); err != nil {
					return errors.Wrap(err, "create dir tree")
				}

				// Create the node
				if _, err := s.client.Create(key, value, flags, zk.WorldACL(zk.PermAll)); err != nil {
					// Node exist error (when previous nil)
					if err == zk.ErrNodeExists {
						return ErrKeyExists
					}
					return err
				}
			} else {
				// Node Exists error (when previous nil)
				if err == zk.ErrNodeExists {
					return ErrKeyExists
				}

				// Unhandled error
				return err
			}
		}

	}

	return nil
}

// AtomicDelete deletes a value at "key" if the key
// has not been modified in the meantime, throws an
// error if this is the case
func (s *ClientImpl) AtomicDelete(key string, previous *KVPair) (bool, error) {
	err := s.client.Delete(key, int32(previous.Stat.Version))
	if err != nil {
		// Key not found
		if err == zk.ErrNoNode {
			return false, ErrKeyNotFound
		}
		// Compare failed
		if err == zk.ErrBadVersion {
			return false, ErrKeyModified
		}
		// General store error
		return false, err
	}
	return true, nil
}

func (s *ClientImpl) CreateDir(key string) error {
	exists, err := s.Exists(key)
	if err != nil {
		return err
	} else if !exists {
		return s.createFullPath(SplitPath(strings.TrimSuffix(key, "/")), false)
	} else {
		return nil
	}
}

// Close closes the ClientImpl connection
func (s *ClientImpl) Close() {
	s.client.Close()
}
