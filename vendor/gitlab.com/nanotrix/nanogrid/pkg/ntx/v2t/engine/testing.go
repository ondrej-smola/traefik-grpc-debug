package engine

import (
	"time"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/timeutil"
	"golang.org/x/net/context"
)

type StreamOrError struct {
	Stream *EngineStream
	Err    error
}

func TestLabelEvent() *Event {
	ev := &Event{Body: &Event_Label_{Label: &Event_Label{Label: &Event_Label_Item{Item: "label"}}}}

	if err := ev.Valid(); err != nil {
		panic(errors.Wrap(err, "Invalid label event"))
	}

	return ev
}

func TestTimestampEvent() *Event {
	ev := &Event{Body: &Event_Timestamp_{Timestamp: &Event_Timestamp{Value: &Event_Timestamp_Timestamp{Timestamp: 0}}}}

	if err := ev.Valid(); err != nil {
		panic(errors.Wrap(err, "Invalid timestamp event"))
	}

	return ev
}

func TestAudioEvent() *Event {
	ev := &Event{
		Body: &Event_Audio_{Audio: &Event_Audio{
			Body:     []byte{1, 2, 3, 4},
			Duration: uint64(timeutil.DurationToTicks(1 * time.Second))}},
	}

	if err := ev.Valid(); err != nil {
		panic(errors.Wrap(err, "Invalid audio event"))
	}

	return ev
}

func TestEngineContext() *EngineContext {
	return &EngineContext{
		AudioFormat: &AudioFormat{Formats: &AudioFormat_Auto{Auto: &AudioFormat_AutoDetect{}}},
		Config:      &EngineContext_Vad{Vad: &EngineContext_VADConfig{}},
	}
}

func TestContextStart() *EngineContextStart {
	return &EngineContextStart{
		Context: TestEngineContext(),
	}
}

func TestStartStream() *EngineStream {
	return &EngineStream{Payload: &EngineStream_Start{Start: &EngineContextStart{}}}
}

type TestEngineServer struct {
	StreamingRecognizeIn  chan EngineService_StreamingRecognizeServer
	StreamingRecognizeOut chan error
	HealthCheckIn         chan *ntx.ServiceNameWithContext
	HealthCheckOut        chan *ntx.ServingStatusOrError
}

func (t *TestEngineServer) StreamingRecognize(serv EngineService_StreamingRecognizeServer) error {
	t.StreamingRecognizeIn <- serv
	return <-t.StreamingRecognizeOut
}

func (s *TestEngineServer) Check(ctx context.Context, req *ntx.HealthCheckRequest) (*ntx.HealthCheckResponse, error) {
	s.HealthCheckIn <- &ntx.ServiceNameWithContext{Service: req.Service, Context: ctx}

	select {
	case resp := <-s.HealthCheckOut:
		return &ntx.HealthCheckResponse{Status: resp.Status}, resp.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *TestEngineServer) AutoHealthCheck() {
	go func() {
		for range s.HealthCheckIn {
			s.HealthCheckOut <- &ntx.ServingStatusOrError{Status: ntx.HealthCheckResponse_SERVING}
		}
	}()
}

func NewTestEngineServer() *TestEngineServer {
	return &TestEngineServer{
		StreamingRecognizeIn:  make(chan EngineService_StreamingRecognizeServer),
		StreamingRecognizeOut: make(chan error),
		HealthCheckIn:         make(chan *ntx.ServiceNameWithContext),
		HealthCheckOut:        make(chan *ntx.ServingStatusOrError),
	}
}

func Ts(o uint64) *Event {
	return TimestampEvent(o)
}
func Rts(o uint64) *Event {
	return RecoveryTimestampEvent(o)
}

func Lab(s string) *Event {
	return LabelItemEvent(s)
}

func Aud(offset, duration uint64, body []byte) *Event {
	a := AudioEvent(body)
	b := a.Body.(*Event_Audio_)
	b.Audio.Duration = duration
	b.Audio.Offset = offset
	return a
}

type EventsOrError struct {
	Events []*Event
	Err    error
}

type TestEventReadCloser struct {
	ReadIn   chan bool
	ReadOut  chan *EventsOrError
	CloseIn  chan bool
	CloseOut chan error
}

func NewTestEventReadCloser() *TestEventReadCloser {
	return &TestEventReadCloser{
		ReadIn:   make(chan bool),
		ReadOut:  make(chan *EventsOrError),
		CloseIn:  make(chan bool),
		CloseOut: make(chan error),
	}
}

func (t *TestEventReadCloser) Read() ([]*Event, error) {
	t.ReadIn <- true
	res := <-t.ReadOut
	return res.Events, res.Err
}

func (t *TestEventReadCloser) Close() error {
	t.CloseIn <- true
	return <-t.CloseOut
}
