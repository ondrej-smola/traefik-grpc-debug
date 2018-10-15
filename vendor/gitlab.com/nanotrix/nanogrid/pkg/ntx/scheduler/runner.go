package scheduler

import (
	"context"

	"github.com/pkg/errors"
)

type (
	blockTillStarted struct {
		next  Runner
		block func(s TaskState) bool
	}

	noopRunner struct {
		next Runner
	}
)

func NoopRunner(next Runner) Runner {
	return &noopRunner{next: next}
}

func BlockWhileNotStarted(next Runner) Runner {
	return BlockWhileRunner(next, func(s TaskState) bool {
		return !s.IsStarted()
	})
}

func BlockWhileRunner(next Runner, block func(s TaskState) bool) Runner {
	return &blockTillStarted{
		next:  next,
		block: block,
	}
}

func (r *blockTillStarted) Run(req Request, ctx context.Context) (Task, error) {
	notifyStarted := make(chan bool, 1)

	nextReq := RequestWithCallback(req, func(st TaskState) {
		if !r.block(st) {
			select {
			case notifyStarted <- true:
			default:
			}
		}
	})

	t, err := r.next.Run(nextReq, ctx)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		t.Release()
		return nil, ctx.Err()
	case <-notifyStarted:
		snap := t.Snapshot()
		if snap.State.IsTerminal() {
			t.Kill()
			return nil, errors.Errorf("Task terminated: %v", snap.State)
		} else {
			return t, nil
		}
	}
}

func (r *noopRunner) Run(req Request, ctx context.Context) (Task, error) {
	return r.next.Run(req, ctx)
}
