package engine

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
)

type (
	EventReader interface {
		Read() ([]*Event, error)
	}

	EventReaderFunc func() ([]*Event, error)

	EventReadCloser interface {
		EventReader
		io.Closer
	}

	EventWriter interface {
		Write(...*Event) error
	}

	EventWriterFunc func(...*Event) error

	EventWriteCloser interface {
		EventWriter
		io.Closer
	}

	binaryEventReader struct {
		from io.ReadCloser
	}

	jsonEventReader struct {
		from  *bufio.Scanner
		close io.Closer
	}

	jsonEventWriter struct {
		to      io.WriteCloser
		marshal jsonpb.Marshaler
	}
)

func (c EventReaderFunc) Read() ([]*Event, error) {
	return c()
}

func (c EventWriterFunc) Write(e ...*Event) error {
	return c(e...)
}

func AudioEventReader(r io.ReadCloser) EventReadCloser {
	return &binaryEventReader{from: r}
}

func JsonEventWriter(r io.WriteCloser) EventWriteCloser {
	return &jsonEventWriter{to: r, marshal: jsonpb.Marshaler{EmitDefaults: true}}
}

func JsonEventReader(r io.ReadCloser) EventReadCloser {
	return &jsonEventReader{from: bufio.NewScanner(r), close: r}
}

func (b *binaryEventReader) Read() ([]*Event, error) {
	to := make([]byte, 4096)
	n, err := b.from.Read(to)
	return []*Event{AudioEvent(to[:n])}, err
}

func (b *binaryEventReader) Close() error {
	return b.from.Close()
}

func (b *jsonEventReader) Read() ([]*Event, error) {
	if b.from.Scan() {
		ev := &Event{}
		return []*Event{ev}, errors.Wrap(jsonpb.UnmarshalString(b.from.Text(), ev), "Event unmarshalling failed")
	} else if err := b.from.Err(); err != nil {
		return nil, err
	} else {
		return nil, io.EOF
	}
}

func (b *jsonEventReader) Close() error {
	return b.close.Close()
}

func (j *jsonEventWriter) Write(evs ...*Event) error {
	for _, e := range evs {
		b, err := j.marshal.MarshalToString(e)
		if err != nil {
			return err
		}
		if _, err = j.to.Write([]byte(b + "\n")); err != nil {
			return err
		}
	}
	return nil
}

func (b *jsonEventWriter) Close() error {
	return b.to.Close()
}

func (e *Event) IsRecoveryTimestamp() (offset uint64, is bool) {
	switch b := e.Body.(type) {
	case *Event_Timestamp_:
		switch v := b.Timestamp.Value.(type) {
		case *Event_Timestamp_Recovery:
			is, offset = true, v.Recovery
		}
	}
	return
}

func (e *Event) AddOffset(offset uint64) {
	switch b := e.Body.(type) {

	// only add offset to timestamp events
	case *Event_Timestamp_:
		switch v := b.Timestamp.Value.(type) {
		case *Event_Timestamp_Recovery:
			v.Recovery += offset
		case *Event_Timestamp_Timestamp:
			v.Timestamp += offset
		}
	}
}

func (e *Event) IsTimestamp() (offset uint64, is bool) {
	switch b := e.Body.(type) {
	case *Event_Timestamp_:
		switch v := b.Timestamp.Value.(type) {
		case *Event_Timestamp_Timestamp:
			is, offset = true, v.Timestamp
		}
	}
	return
}

func (e *Event) IsAudio() (offset uint64, is bool) {
	switch b := e.Body.(type) {
	case *Event_Audio_:
		is, offset = true, b.Audio.Offset
	}
	return
}

func (e *Event) HasOffset() (offset uint64, has bool) {
	offset, has = e.IsRecoveryTimestamp()
	if !has {
		offset, has = e.IsTimestamp()
	}

	if !has {
		offset, has = e.IsAudio()
	}

	return
}

func (e *Event) ToText() string {
	switch b := e.Body.(type) {

	case *Event_Label_:
		switch e := b.Label.Label.(type) {
		case *Event_Label_Item:
			return e.Item
		case *Event_Label_Plus:
			return e.Plus
		}
	}

	return ""
}

type (
	EventWithOffset struct {
		Event  *Event
		Offset uint64
	}

	EventsSeq []*EventWithOffset

	EventsByOffsetSelectFn func(offset uint64) bool
)

func NewEventsSeq(startOffset uint64, evs ...*Event) EventsSeq {
	res := make(EventsSeq, len(evs))
	current := startOffset

	for i, e := range evs {
		if o, has := e.HasOffset(); has {
			current = o
		}
		res[i] = &EventWithOffset{
			Event:  e,
			Offset: current,
		}
	}

	return res
}

func (s EventsSeq) FirstRecoveryOffset() (offset uint64, ok bool) {
	for _, e := range s {
		if offset, ok = e.Event.IsRecoveryTimestamp(); ok {
			break
		}
	}
	return
}

func (s EventsSeq) LastOffset() uint64 {
	if len(s) == 0 {
		return 0
	}

	return s[len(s)-1].Offset
}

func (s EventsSeq) LastRecoveryOffset() (offset uint64, ok bool) {
	for _, e := range s {
		if o, is := e.Event.IsRecoveryTimestamp(); is {
			offset, ok = o, is
		}
	}
	return
}

func (s EventsSeq) AddOffset(offset uint64) {
	for _, e := range s {
		e.Offset += offset
		e.Event.AddOffset(offset)
	}
}

func (e EventsSeq) Clone() EventsSeq {
	clone := make(EventsSeq, len(e))

	for i, e := range e {
		clone[i] = &EventWithOffset{
			Event:  e.Event.Clone(),
			Offset: e.Offset,
		}
	}

	return clone
}

func (e *EventWithOffset) Equal(other *EventWithOffset) bool {
	if e == nil && other == nil {
		return true
	}

	return e.Offset == other.Offset && proto.Equal(e.Event, other.Event)
}

func (e EventsSeq) Equal(other EventsSeq) bool {
	if len(e) != len(other) {
		return false
	}

	for i, ev := range e {
		if !ev.Equal(other[i]) {
			return false
		}
	}

	return true
}

func (e EventsSeq) Events() []*Event {
	res := make([]*Event, len(e))
	for i, ev := range e {
		res[i] = ev.Event
	}

	return res
}

func (e EventsSeq) Size() int {
	return len(e)
}

func (e EventsSeq) Empty() bool {
	return len(e) == 0
}

func (e EventsSeq) SelectInRange(from uint64, to uint64) EventsSeq {
	return e.SelectByOffset(func(o uint64) bool {
		return o >= from && o < to
	})
}

func (e EventsSeq) SelectFrom(from uint64) EventsSeq {
	return e.SelectByOffset(func(o uint64) bool {
		return o >= from
	})
}

func (e EventsSeq) SelectTo(to uint64) EventsSeq {
	return e.SelectByOffset(func(o uint64) bool {
		return o < to
	})
}

func (e EventsSeq) SelectByOffset(a EventsByOffsetSelectFn) EventsSeq {
	res := EventsSeq{}

	for _, ev := range e {
		if a(ev.Offset) {
			res = append(res, ev)
		}
	}

	return res
}

func (e *EventWithOffset) String() string {
	return fmt.Sprintf("EventWithOffset[offset:%v,event:%v]", e.Offset, e.Event)
}

func (e EventsSeq) String() string {
	res := make([]string, len(e))

	for i, ev := range e {
		res[i] = ev.String()
	}

	return strings.Join(res, ",")
}

func AudioEvent(bytes []byte) *Event {
	return &Event{
		Body: &Event_Audio_{
			Audio: &Event_Audio{Body: bytes},
		},
	}
}

func LabelItemEvent(item string) *Event {
	return &Event{
		Body: &Event_Label_{
			Label: &Event_Label{
				Label: &Event_Label_Item{
					Item: item,
				},
			},
		},
	}
}

func LabelPlusEvent(plus string) *Event {
	return &Event{
		Body: &Event_Label_{
			Label: &Event_Label{
				Label: &Event_Label_Plus{
					Plus: plus,
				},
			},
		},
	}
}

func TimestampEvent(ts uint64) *Event {
	return &Event{
		Body: &Event_Timestamp_{
			Timestamp: &Event_Timestamp{
				Value: &Event_Timestamp_Timestamp{
					Timestamp: ts,
				},
			},
		},
	}
}

func RecoveryTimestampEvent(rts uint64) *Event {
	return &Event{
		Body: &Event_Timestamp_{
			Timestamp: &Event_Timestamp{
				Value: &Event_Timestamp_Recovery{
					Recovery: rts,
				},
			},
		},
	}
}
