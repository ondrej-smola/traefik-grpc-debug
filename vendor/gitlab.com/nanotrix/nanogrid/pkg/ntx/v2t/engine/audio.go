package engine

import "github.com/pkg/errors"

type (
	AudioEvents []*Event_Audio
)

func EventsToAudioEvents(evs []*Event) (AudioEvents, error) {
	res := make(AudioEvents, len(evs))
	for i, e := range evs {
		switch t := e.Body.(type) {
		case *Event_Audio_:
			if t.Audio != nil {
				res[i] = t.Audio
			} else {
				return nil, errors.Errorf("Event[%v] - audio not set on audio event", i)
			}
		default:
			return nil, errors.Errorf("Event[%v] - not at audio event, got %T", i, e.Body)
		}
	}

	return res, nil
}

func (a AudioEvents) Empty() bool {
	return len(a) == 0
}

func (a AudioEvents) Offset() uint64 {
	if a.Empty() {
		return 0
	}
	return a[0].Offset
}

func (a AudioEvents) End() uint64 {
	if a.Empty() {
		return 0
	}

	last := a[len(a)-1]
	return last.Offset + last.Duration
}

func (a AudioEvents) Continuous() bool {
	prevEnd := uint64(0)

	for i, e := range a {
		if i == 0 || prevEnd == e.Offset {
			prevEnd = e.Offset + e.Duration
		} else {
			return false
		}
	}

	return true
}
