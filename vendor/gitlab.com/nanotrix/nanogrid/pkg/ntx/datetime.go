package ntx

import (
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
)

func (p *TimeOffsetFraction) Valid() error {
	if p == nil {
		return errors.New("TimeOffsetFraction: is nil")
	}

	return nil
}

func (p *TimeOffsetFraction) ToTicks() uint64 {
	if p == nil {
		return 0
	} else {
		return p.Seconds*10*1000*1000 + uint64(p.Ticks)
	}
}

func TimeOffsetFractionFromTicks(ticks uint64) *TimeOffsetFraction {
	secs := ticks / timeutil.TicksPerSecond
	return &TimeOffsetFraction{
		Seconds: secs,
		Ticks:   uint32(ticks % timeutil.TicksPerSecond),
	}
}
