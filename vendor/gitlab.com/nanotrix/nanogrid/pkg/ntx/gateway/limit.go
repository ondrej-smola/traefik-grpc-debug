package gateway

import (
	"fmt"

	"github.com/pkg/errors"
)

type LimitError struct {
	msg string
}

func NewLimitError(msg string) LimitError {
	return LimitError{msg}
}

func NewLimitErrorf(format string, args ...interface{}) LimitError {
	return LimitError{fmt.Sprintf(format, args...)}
}

func IsLimitError(e error) bool {
	if e == nil {
		return false
	}

	e = errors.Cause(e)
	_, ok := e.(LimitError)
	return ok
}

func (e LimitError) Error() string {
	return e.msg
}

func (q *RoleLimit) Valid() error {
	if q == nil {
		return errors.New("is nil")
	}

	if q.Role == "" {
		return errors.New("role: is blank")
	}

	if q.MaxTasks == 0 {
		return errors.New("maxTasks: must be > 0")
	}

	for i, g := range q.Gateways {
		if err := g.Valid(); err != nil {
			return errors.Wrapf(err, "gateways[%v]", i)
		}
	}

	return nil
}

func (q *RoleLimit) TasksCount() int {
	total := 0

	for _, g := range q.Gateways {
		total += len(g.Tasks)
	}

	return total
}

func (q *RoleLimit) TotalCpu() float64 {
	total := float64(0)

	for _, g := range q.Gateways {
		for _, t := range g.Tasks {
			total += t.Cpus
		}
	}

	return total
}

func (q *RoleLimit) TotalMem() float64 {
	total := float64(0)

	for _, g := range q.Gateways {
		for _, t := range g.Tasks {
			total += t.Mem
		}
	}

	return total
}

func (q *RoleLimits) Valid() error {
	if q == nil {
		return errors.New("is nil")
	}

	for i, r := range q.Roles {
		if err := r.Valid(); err != nil {
			return errors.Wrapf(err, "roles[%v]", i)
		}
	}

	return nil
}

func (q *RoleLimit_Gateway) Valid() error {
	if q == nil {
		return errors.New("is nil")
	}

	if q.Id == "" {
		return errors.New("id: is blank")
	}

	for i, t := range q.Tasks {
		if err := t.Valid(); err != nil {
			return errors.Wrapf(err, "tasks[%v]", i)
		}
	}

	return nil
}

func (q *RoleLimit_Gateway_Task) Valid() error {
	if q == nil {
		return errors.New("is nil")
	}

	if q.ReqId == "" {
		return errors.New("reqId: is blank")
	}

	if q.Cpus <= 0 {
		return errors.Errorf("cpus: must be > 0 (got %v)", q.Cpus)
	}

	if q.Mem <= 0 {
		return errors.Errorf("mem: must be > 0 (got %v)", q.Cpus)
	}

	return nil
}
