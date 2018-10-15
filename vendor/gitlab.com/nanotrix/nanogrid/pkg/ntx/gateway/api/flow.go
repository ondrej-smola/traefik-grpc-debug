package api

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/gateway"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/v2t/engine"
)

type PushOnlySrc struct {
	prev gateway.Src
	ctx  context.Context

	readNext chan bool
}

var _ = gateway.Src(&PushOnlySrc{})

func NewPushOnlySrc(prev gateway.Src, ctx context.Context) *PushOnlySrc {
	readNext := make(chan bool, 1)
	readNext <- true

	return &PushOnlySrc{
		prev:     prev,
		ctx:      ctx,
		readNext: readNext,
	}
}

func (r *PushOnlySrc) RecvMsg() (*engine.EngineStream, error) {
	select {
	case readNext := <-r.readNext:
		if !readNext {
			return engine.PullStream(), nil
		}
	case <-r.ctx.Done():
		return nil, r.ctx.Err()
	}

	m, err := r.prev.RecvMsg()
	if err != nil {
		return nil, err
	}

	switch m.Payload.(type) {
	case *engine.EngineStream_Pull:
		return nil, errors.Errorf("PULL in non flow control")
	default:
		return m, nil
	}

}

func (r *PushOnlySrc) SendMsg(m *engine.EngineStream) error {
	notify := func(readNext bool, msg string, send bool) error {
		select {
		case r.readNext <- readNext:
			if send {
				return r.prev.SendMsg(m)
			} else {
				return nil
			}
		default:
			return errors.Errorf("reader: unexpected %v", msg)
		}
	}

	switch m.Payload.(type) {
	case *engine.EngineStream_Pull:
		return notify(true, "PULL", false)
	case *engine.EngineStream_Push:
		return notify(false, "PUSH", true)
	case *engine.EngineStream_Start:
		return notify(false, "START", true)
	case *engine.EngineStream_End:
		return notify(true, "START", true)
	default:
		return errors.Errorf("unexpected message payload type: %T", m.Payload)
	}
}

func (r *PushOnlySrc) SendHeaders(headers map[string]string) error {
	return r.prev.SendHeaders(headers)
}
