package cache

import (
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/scheduler"
)

type (
	Resources struct {
		Cpus, Mem float64
	}
)

func (r Resources) String() string {
	return fmt.Sprintf("cpus:%v,mem:%v", r.Cpus, r.Mem)
}

func (r *Resources) Valid() error {
	if r == nil {
		return errors.New("is nil")
	}

	if r.Cpus < 0 {
		return errors.Errorf("cpus < 0")
	}

	if r.Mem < 0 {
		return errors.Errorf("mem < 0")
	}

	return nil
}

func (r *Resources) Clone() *Resources {
	return &Resources{
		Cpus: r.Cpus,
		Mem:  r.Mem,
	}
}

func ResourcesFromTaskConfig(cfg *scheduler.TaskConfiguration) *Resources {
	return &Resources{
		Cpus: cfg.Cpus,
		Mem:  cfg.Mem,
	}
}
