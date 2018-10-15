package scheduler

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

type (
	TestRequestWithContext struct {
		Req Request
		Ctx context.Context
	}

	TestRunningTaskOrError struct {
		Task Task
		Err  error
	}

	TestRunner struct {
		RunIn  chan *TestRequestWithContext
		RunOut chan *TestRunningTaskOrError
	}
)

func NewTestRunner() *TestRunner {
	return &TestRunner{
		RunIn:  make(chan *TestRequestWithContext),
		RunOut: make(chan *TestRunningTaskOrError),
	}
}

func (r *TestRunner) Run(req Request, ctx context.Context) (Task, error) {
	r.RunIn <- &TestRequestWithContext{Req: req, Ctx: ctx}
	res := <-r.RunOut
	return res.Task, res.Err
}

func TestTaskConfiguration() *TaskConfiguration {
	cfg := &TaskConfiguration{
		Image: "image:" + util.RandString(8),
		Cpus:  0.1,
		Mem:   61,
	}

	if err := cfg.Valid(); err != nil {
		panic(errors.Wrap(err, "Invalid mock task configuration"))
	}

	return cfg
}

func TestTaskConfigurationWithImage(img string) *TaskConfiguration {
	cfg := TestTaskConfiguration()
	cfg.Image = img
	return cfg
}

type (
	TestTask struct {
		cfg      *TaskConfiguration
		id       string
		endpoint string
		released bool
		killed   bool
		state    TaskState
		sync.Mutex
	}

	TestRequest struct {
		Cfg  TaskConfiguration
		Role string
	}
)

func NewTestTask(tid string, endpoint ...string) *TestTask {
	end := ""
	if len(endpoint) > 0 {
		end = endpoint[0]
	}

	return &TestTask{
		id:       tid,
		endpoint: end,
		cfg:      TestTaskConfiguration(),
		state:    TaskState_RUNNING,
	}
}
func (r *TestTask) SetState(st TaskState) {
	r.Lock()
	r.state = st
	r.Unlock()
}

func (r *TestTask) Snapshot() *TaskSnapshot {
	r.Lock()
	defer r.Unlock()
	return &TaskSnapshot{
		Id:       r.id,
		Endpoint: r.endpoint,
		State:    r.state,
		Config:   r.cfg.Clone(),
	}
}

func (r *TestTask) Id() string {
	return r.id
}

func (r *TestTask) Endpoint() string {
	return r.endpoint
}

func (r *TestTask) SetConfig(cfg *TaskConfiguration) {
	r.cfg = cfg.Clone()
}

func (r *TestTask) Config() *TaskConfiguration {
	return r.cfg.Clone()
}

func (r *TestTask) Release() {
	r.Lock()
	r.state = TaskState_KILLED
	r.released = true
	r.Unlock()
}

func (r *TestTask) IsReleased() bool {
	r.Lock()
	defer r.Unlock()
	return r.released
}

func (r *TestTask) Kill() {
	r.Lock()
	r.state = TaskState_KILLED
	r.killed = true
	r.Unlock()
}

func (r *TestTask) IsKilled() bool {
	r.Lock()
	defer r.Unlock()
	return r.killed
}
