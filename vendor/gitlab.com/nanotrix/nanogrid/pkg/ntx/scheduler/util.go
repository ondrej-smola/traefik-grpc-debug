package scheduler

import (
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
)

func (c *TaskConfiguration) HasImage() bool {
	return c.Image != ""
}

func (t *TaskConfiguration) HasPort() bool {
	return t.Port != 0
}

func (c *TaskConfiguration) Valid() error {
	if c == nil {
		return errors.New("is nil")
	}

	if c.Image == "" {
		return errors.New("image: blank")
	}

	if c.Mem <= 0 {
		return fmt.Errorf("mem: must be > 0 (is '%v')", c.Mem)
	}

	if c.Cpus <= 0 {
		return fmt.Errorf("cpus: must be > 0 (is '%v')", c.Cpus)
	}

	return nil
}

func (c *TaskConfiguration) Clone() *TaskConfiguration {
	return &TaskConfiguration{
		Image: c.Image,
		Cpus:  c.Cpus,
		Mem:   c.Mem,
		Cmd:   c.Cmd,
		Args:  util.CloneStringSlice(c.Args),
		Port:  c.Port,
		Env:   util.CloneStringMap(c.Env),
	}
}

func (st TaskState) IsStarting() bool {
	return st == TaskState_STARTING
}

func (st TaskState) IsStarted() bool {
	return !st.IsWaiting() && !st.IsStarting()
}

func (st TaskState) IsRunning() bool {
	return st == TaskState_RUNNING
}

func (st TaskState) IsWaiting() bool {
	return st == TaskState_WAITING
}

func (st TaskState) IsTerminal() bool {
	return !st.IsWaiting() && !st.IsStarting() && !st.IsRunning()
}

func (st TaskState) Valid() error {
	_, found := TaskState_name[int32(st)]
	if !found {
		return errors.Errorf("Task_State - unknown value %v", st)
	}

	return nil
}
