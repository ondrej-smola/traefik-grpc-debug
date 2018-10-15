package gateway

import (
	"sync"

	"context"

	"github.com/go-kit/kit/log"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth/perm"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/cluster"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/zk"
)

const (
	LimiterZkDefaultPath = "/ntx/gateway/roles"
)

type (
	RoleLimiter interface {
		Allow(reqId string, tkn *auth.NtxToken_Task) error
		Done(reqId string)
	}

	// should return LimitErr when limit reached, nil to allow
	RoleLimitFn func(q *RoleLimit, task *RoleLimit_Gateway_Task) error

	RoleLimiterManagement interface {
		List() ([]*RoleLimit, error)
		Set(limit *RoleLimit) error
		Delete(limit *RoleLimit) error
	}

	zkTask struct {
		role string
		RoleLimit_Gateway_Task
	}

	ZkRoleLimiter struct {
		gwId                     string
		baseDir                  string
		zkConflictMaxAttempts    uint64
		zkConflictMaxSleepMicros uint64

		limitFn     RoleLimitFn
		membersList cluster.Listing

		log      log.Logger
		zkClient zk.Client

		sync.Mutex
		// reqId -> zkTask
		tasks       map[string]*zkTask
		perRoleLock map[string]*sync.Mutex
	}

	NoopRoleLimiter struct{}

	NoopRoleLimiterManagement struct {
	}

	ZkRolLimiterOption func(z *ZkRoleLimiter)
)

var ErrTooManyConflicts = errors.New("zk: too many modified conflicts")

func WithZkLimiterGatewayId(id string) ZkRolLimiterOption {
	return func(z *ZkRoleLimiter) {
		z.gwId = id
	}
}

func WithZkLimiterBaseDir(baseDir string) ZkRolLimiterOption {
	return func(z *ZkRoleLimiter) {
		z.baseDir = baseDir
	}
}

func WithZkConflictMaxAttempts(attempts uint64) ZkRolLimiterOption {
	if attempts == 0 {
		panic("attempts must be > 0")
	}

	return func(z *ZkRoleLimiter) {
		z.zkConflictMaxAttempts = attempts
	}
}

func WithZkLimiterLogger(log log.Logger) ZkRolLimiterOption {
	return func(z *ZkRoleLimiter) {
		z.log = log
	}
}

func WithZkLimiterLimitFn(fn RoleLimitFn) ZkRolLimiterOption {
	return func(z *ZkRoleLimiter) {
		z.limitFn = fn
	}
}

// wrap gateway DoFn with RoleLimiter
func LimiterFn(keeper RoleLimiter, next DoFn) DoFn {
	return func(reqId string, from Src, to DestFn, tkn *auth.NtxToken, ctx context.Context) error {
		if tkn.Task == nil {
			return errors.New("limiter: token: task: not set")
		}

		if err := keeper.Allow(reqId, tkn.Task); err != nil {
			return err
		}
		defer keeper.Done(reqId)

		return next(reqId, from, to, tkn, ctx)
	}
}

func NewZkLimiter(client zk.Client, listing cluster.Listing, opts ...ZkRolLimiterOption) *ZkRoleLimiter {
	kp := &ZkRoleLimiter{
		gwId:                  "gateway",
		baseDir:               LimiterZkDefaultPath,
		zkConflictMaxAttempts: 32,
		zkClient:              client,
		tasks:                 map[string]*zkTask{},

		limitFn:     DefaultLimitFn,
		membersList: listing,
		perRoleLock: map[string]*sync.Mutex{},

		log: log.NewNopLogger(),
	}

	for _, o := range opts {
		o(kp)
	}

	return kp
}

func (k *ZkRoleLimiter) getPathForRole(role string) string {
	return zk.Join(k.baseDir, role)
}

func (k *ZkRoleLimiter) Set(limit *RoleLimit) error {
	if limit.Role == perm.UnlimitedRole() {
		return errors.Errorf("reserved role name: %v", limit.Role)
	}

	path := k.getPathForRole(limit.Role)
	oldLimit, prevKv, err := k.getLimitForPath(path)
	if err != nil {
		if err != zk.ErrKeyNotFound {
			return err
		}
	}

	merged := proto.Clone(limit).(*RoleLimit)
	if oldLimit != nil {
		merged.Gateways = oldLimit.Gateways
		k.setTasksForSelf(merged, k.getActiveTasksForRole(limit.Role))
	}

	for i := uint64(0); i < k.zkConflictMaxAttempts; i++ {
		if err := k.writeRoleLimit(path, merged, prevKv); err != nil {
			if err == zk.ErrKeyModified {
				continue
			} else {
				return err
			}
		} else {
			return nil
		}
	}

	return ErrTooManyConflicts
}

func (k *ZkRoleLimiter) Delete(limit *RoleLimit) error {
	path := k.getPathForRole(limit.Role)
	err := k.zkClient.Delete(path)
	if err == zk.ErrKeyNotFound {
		err = nil
	}

	return errors.Wrapf(err, "path %v", path)
}

func (k *ZkRoleLimiter) List() ([]*RoleLimit, error) {
	kvs, err := k.zkClient.List(k.baseDir)

	if err != nil {
		if err == zk.ErrKeyNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	limits := make([]*RoleLimit, len(kvs))

	for i, kv := range kvs {
		limit := &RoleLimit{}
		if err := proto.Unmarshal(kv.Value, limit); err != nil {
			return nil, errors.Wrap(err, "proto")
		} else if err := limit.Valid(); err != nil {
			return nil, errors.Wrap(err, "roleLimit")
		} else {
			limits[i] = limit
		}
	}

	return limits, nil
}

func (k *ZkRoleLimiter) getLockForRole(role string) *sync.Mutex {
	k.Lock()
	defer k.Unlock()

	lck, ok := k.perRoleLock[role]
	if ok {
		return lck
	}

	lck = &sync.Mutex{}
	k.perRoleLock[role] = lck

	return lck
}

func (k *ZkRoleLimiter) Done(reqId string) {
	k.Lock()
	task, ok := k.tasks[reqId]
	delete(k.tasks, reqId)
	k.Unlock()

	if !ok {
		return
	}

	lck := k.getLockForRole(task.role)
	lck.Lock()
	defer lck.Unlock()
	for i := uint64(0); i < k.zkConflictMaxAttempts; i++ {
		path := k.getPathForRole(task.role)
		roleLimit, kv, err := k.getLimitForPath(path)
		if err != nil {
			k.log.Log("ev", "req_done", "req", reqId, "cause", "zk_get_limit", "path", path, "err", err)
			return
		}

		k.removeInactiveMembers(roleLimit)
		k.setTasksForSelf(roleLimit, k.getActiveTasksForRole(task.role))

		if err := k.writeRoleLimit(path, roleLimit, kv); err != nil {
			if err == zk.ErrKeyModified {
				k.log.Log("ev", "zk_key_modified", "path", path, "attempt", i, "debug", true)
				continue
			} else {
				k.log.Log("ev", "req_done", "req", reqId, "cause", "zk_write_limit", "path", path, "err", err)
			}
		}
		break
	}
}

func (k *ZkRoleLimiter) getLimitForPath(path string) (*RoleLimit, *zk.KVPair, error) {
	kv, err := k.zkClient.Get(path)

	if err != nil {
		return nil, nil, err
	}

	roleLimit := &RoleLimit{}
	if err := proto.Unmarshal(kv.Value, roleLimit); err != nil {
		return nil, nil, errors.Wrap(err, "proto")
	} else if err := roleLimit.Valid(); err != nil {
		return nil, nil, errors.Wrap(err, "roleLimit")
	} else {
		k.removeInactiveMembers(roleLimit)
		return roleLimit, kv, nil
	}
}

func (k *ZkRoleLimiter) removeInactiveMembers(q *RoleLimit) {
	var active []*RoleLimit_Gateway

	members := k.membersList.List()
	for i := range q.Gateways {
		if members.ContainsId(q.Gateways[i].Id) {
			active = append(active, q.Gateways[i])
		}
	}

	q.Gateways = active
}

func (k *ZkRoleLimiter) Allow(reqId string, task *auth.NtxToken_Task) error {
	role := task.Role

	if role == "" {
		return errors.New("token: task: role: blank")
	}

	if role == perm.UnlimitedRole() {
		return nil
	}

	newTask := &zkTask{
		role: role,
		RoleLimit_Gateway_Task: RoleLimit_Gateway_Task{
			ReqId: reqId,
			Cpus:  task.Cfg.Cpus,
			Mem:   task.Cfg.Mem,
		},
	}

	lck := k.getLockForRole(role)
	lck.Lock()
	defer lck.Unlock()

	for i := uint64(0); i < k.zkConflictMaxAttempts; i++ {
		path := k.getPathForRole(role)
		roleLimit, kv, err := k.getLimitForPath(path)
		if err != nil {
			if err == zk.ErrKeyNotFound {
				return nil
			} else {
				return errors.Wrapf(err, "role: %v", role)
			}
		}

		if err := k.limitFn(roleLimit, &newTask.RoleLimit_Gateway_Task); err != nil {
			return err
		}

		active := append(k.getActiveTasksForRole(role), &newTask.RoleLimit_Gateway_Task)
		k.setTasksForSelf(roleLimit, active)

		if err := k.writeRoleLimit(path, roleLimit, kv); err != nil {
			if err == zk.ErrKeyModified {
				k.log.Log("ev", "zk_key_modified", "attempt", i, "path", path, "debug", true)
				continue
			} else {
				return errors.Wrap(err, "zk put")
			}
		} else {
			k.Lock()
			k.tasks[reqId] = newTask
			k.Unlock()
			return nil
		}
	}

	return ErrTooManyConflicts
}

func (k *ZkRoleLimiter) writeRoleLimit(path string, limit *RoleLimit, previous *zk.KVPair) error {
	if res, err := proto.Marshal(limit); err != nil {
		return errors.Wrap(err, "proto")
	} else if err := k.zkClient.AtomicPut(path, res, previous); err != nil {
		return err
	} else {
		return nil
	}
}

func (k *ZkRoleLimiter) getActiveTasksForRole(role string) []*RoleLimit_Gateway_Task {
	var activeTasks []*RoleLimit_Gateway_Task
	k.Lock()
	for _, t := range k.tasks {
		if t.role == role {
			cpy := t
			activeTasks = append(activeTasks, &cpy.RoleLimit_Gateway_Task)
		}
	}

	k.Unlock()
	return activeTasks
}

func (k *ZkRoleLimiter) setTasksForSelf(qos *RoleLimit, tasks []*RoleLimit_Gateway_Task) {
	if len(tasks) == 0 {
		for i, g := range qos.Gateways {
			if g.Id == k.gwId {
				qos.Gateways = append(qos.Gateways[:i], qos.Gateways[i+1:]...)
				g.Tasks = tasks
				return
			}
		}
	} else {
		for _, g := range qos.Gateways {
			if g.Id == k.gwId {
				g.Tasks = tasks
				return
			}
		}

		qos.Gateways = append(
			qos.Gateways,
			&RoleLimit_Gateway{
				Id:    k.gwId,
				Tasks: tasks,
			},
		)
	}
}

func DefaultLimitFn(q *RoleLimit, task *RoleLimit_Gateway_Task) error {
	tasksCount := uint64(q.TasksCount())

	if q.MaxTasks > 0 && q.MaxTasks <= tasksCount {
		return NewLimitErrorf("tasks: requested %v of %v allowed", tasksCount+1, tasksCount)
	}

	if q.MaxCpus > 0 {
		newCpu := q.TotalCpu() + task.Cpus
		if q.MaxCpus < newCpu {
			return NewLimitErrorf("cpu: requested %.2f of %.2f allowed", newCpu, q.MaxCpus)
		}
	}

	if q.MaxMem > 0 {
		newMem := q.TotalMem() + task.Mem
		if q.MaxMem < newMem {
			return NewLimitErrorf("mem: requested %.2f of %.2f allowed", newMem, q.MaxMem)
		}
	}

	return nil
}
