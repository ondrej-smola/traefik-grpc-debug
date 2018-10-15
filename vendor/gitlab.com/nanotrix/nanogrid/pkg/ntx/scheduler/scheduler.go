package scheduler

import (
	"context"
)

type (
	// Task scheduled by some scheduler
	// Must be terminated by calling Release or Kill, after that further calls to Release and Kill are ignored
	Task interface {
		Snapshot() *TaskSnapshot
		// release task so it can be reused/cached
		Release()
		// kill task preventing any reuse
		Kill()
	}

	Runner interface {
		Run(Request, context.Context) (Task, error)
	}

	// Invoked when task state changes, endpoint may be empty
	// Must be set on context when invoking Runner.Run function

	RunnerFunc func(Request, context.Context) (Task, error)

	TaskStateCallback func(s TaskState)

	Request interface {
		Config() *TaskConfiguration
		OnStateChange(s TaskState)
	}

	request struct {
		cfg *TaskConfiguration
		cb  TaskStateCallback
	}
)

func (f RunnerFunc) Run(req Request, ctx context.Context) (Task, error) {
	return f(req, ctx)
}

func NewRequest(cfg *TaskConfiguration) Request {
	return &request{cfg: cfg}
}

func RequestWithCallback(prev Request, cb TaskStateCallback) Request {
	return &request{
		cfg: prev.Config(),
		cb: func(s TaskState) {
			prev.OnStateChange(s)
			cb(s)
		},
	}
}

func (r *request) Config() *TaskConfiguration {
	return r.cfg
}

func (r *request) OnStateChange(s TaskState) {
	if r.cb != nil {
		r.cb(s)
	}
}
