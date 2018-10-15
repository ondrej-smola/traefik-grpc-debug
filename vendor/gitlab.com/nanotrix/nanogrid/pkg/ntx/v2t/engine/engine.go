package engine

import (
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

func (e *Events) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	for i, ev := range e.Events {
		if err := ev.Valid(); err != nil {
			return errors.Wrapf(err, "event[%v]", i)
		}
	}

	return nil
}

func (e *Event) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	if e.Body == nil {
		return errors.New("body: is nil")
	}

	if t := e.GetTimestamp(); t != nil {
		if err := t.Valid(); err != nil {
			return errors.Wrap(err, "body: timestamp")
		}
	}

	if t := e.GetLabel(); t != nil {
		if err := t.Valid(); err != nil {
			return errors.Wrap(err, "body: label")
		}
	}

	if t := e.GetAudio(); t != nil {
		if err := t.Valid(); err != nil {
			return errors.Wrap(err, "body: audio")
		}
	}

	if t := e.GetMeta(); t != nil {
		if err := t.Valid(); err != nil {
			return errors.Wrap(err, "body: timestamp")
		}
	}

	return nil
}

func (e *Event) Clone() *Event {
	return proto.Clone(e).(*Event)
}

func (e *EngineContext) Clone() *EngineContext {
	return proto.Clone(e).(*EngineContext)
}

func (e *EngineStream) Clone() *EngineStream {
	return proto.Clone(e).(*EngineStream)
}

func (e *AudioFormat) SetProbeSizeBytes(size uint32) {
	switch f := e.Formats.(type) {
	case *AudioFormat_Auto:
		f.Auto.ProbeSizeBytes = size
	}
}

func (m EngineModule) Valid() error {
	if _, found := EngineModule_name[int32(m)]; !found {
		return errors.Errorf("Unknown engine module enum number '%v'", m)
	}

	return nil
}

func (t *Event_Timestamp) Valid() error {
	if t == nil {
		return errors.New("is nil")
	}

	return nil
}

func (l *Event_Label) Valid() error {
	if l == nil {
		return errors.New(" is nil")
	}

	return nil
}

func (b *Event_Audio) Valid() error {
	if b == nil {
		return errors.New("is nil")
	}

	return nil
}

func (c *Event_Meta_Confidence) Valid() error {
	if c == nil {
		return errors.New("is nil")
	}
	return nil
}

func (b *Event_Meta) Valid() error {
	if b == nil {
		return errors.New("is nil")
	}

	if conf := b.GetConfidence(); conf != nil {
		if err := conf.Valid(); err != nil {
			return errors.Wrap(err, "confidence")
		}
	}

	return nil
}

func (l *Lexicon) Valid() error {
	if l == nil {
		return errors.New("is nil")
	}

	for i, e := range l.Items {
		if err := e.Valid(); err != nil {
			return errors.Wrapf(err, "items[%v]", i)
		}
	}

	return nil
}

func (a *AudioFormat_AutoDetect) Valid() error {
	if a == nil {
		return errors.New("is nil")
	}

	return nil
}

func (a *AudioFormat_Auto) Valid() error {
	if a == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(a.Auto.Valid(), "auto")
}

func (a *AudioFormat_Header) Valid() error {
	if a == nil {
		return errors.New("is nil")
	}

	if len(a.Header) == 0 {
		return errors.New("header: is empty")
	}

	return nil
}

func (a *AudioFormat_Header_) Valid() error {
	if a == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(a.Header.Valid(), "header")
}

func (r AudioFormat_SampleRate) Valid() error {
	if _, ok := AudioFormat_SampleRate_name[int32(r)]; !ok {
		return errors.Errorf("unknown sample rate: %v", r)
	}

	return nil
}

func (r AudioFormat_SampleFormat) Valid() error {
	if _, ok := AudioFormat_SampleFormat_name[int32(r)]; !ok {
		return errors.Errorf("unknown sample format: %v", r)
	}

	return nil
}

func (r AudioFormat_ChannelLayout) Valid() error {
	if _, ok := AudioFormat_ChannelLayout_name[int32(r)]; !ok {
		return errors.Errorf("unknown channel layout: %v", r)
	}

	return nil
}

func (a *AudioFormat_Pcm) Valid() error {
	if a == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(a.Pcm.Valid(), "pcm")
}

func (a *AudioFormat_PCM) Valid() error {
	if a == nil {
		return errors.New("is nil")
	}

	if err := a.SampleRate.Valid(); err != nil {
		return errors.Wrap(err, "sample rate")
	}

	if err := a.ChannelLayout.Valid(); err != nil {
		return errors.Wrap(err, "channel layout")
	}

	if err := a.SampleFormat.Valid(); err != nil {
		return errors.Wrap(err, "sample format")
	}

	return nil
}

func (a *AudioFormat) Valid() error {
	if a == nil {
		return errors.New("is nil")
	}

	switch t := a.Formats.(type) {
	case *AudioFormat_Auto:
		if err := t.Valid(); err != nil {
			return errors.Wrap(err, "format")
		}
	case *AudioFormat_Header_:
		if err := t.Valid(); err != nil {
			return errors.Wrap(err, "format")
		}

	case *AudioFormat_Pcm:
		if err := t.Valid(); err != nil {
			return errors.Wrap(err, "format")
		}
	default:
		return errors.Errorf("unknown format: %T", t)
	}

	return nil
}

func (e *EngineContext_VADConfig) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return nil
}

func (e *EngineContext_V2TConfig) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	if e.WithLexicon != nil {
		if err := e.WithLexicon.Valid(); err != nil {
			return errors.Wrap(err, "with lexicon")
		}
	}

	if e.WithPPC != nil {
		if err := e.WithPPC.Valid(); err != nil {
			return errors.Wrap(err, "with ppc")
		}
	}

	if e.WithVAD != nil {
		if err := e.WithVAD.Valid(); err != nil {
			return errors.Wrap(err, "with vad")
		}
	}

	return nil
}

func (e *EngineContext_PPCConfig) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return nil
}

func (e *EngineContext_Vad) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return e.Vad.Valid()
}

func (e *EngineContext_V2T) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return e.V2T.Valid()
}

func (e *EngineContext_Ppc) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return e.Ppc.Valid()
}

func (a EngineContext_AudioChannel) Valid() error {
	if _, ok := EngineContext_AudioChannel_name[int32(a)]; !ok {
		return errors.Errorf("unknown value: %v", a)
	} else {
		return nil
	}
}

func (e *EngineContext) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	if err := e.AudioFormat.Valid(); err != nil {
		return errors.Wrap(err, "audio format")
	}

	if err := e.AudioChannel.Valid(); err != nil {
		return errors.Wrap(err, "audio channel")
	}

	switch t := e.Config.(type) {
	case *EngineContext_Vad:
		return errors.Wrap(t.Valid(), "vad")
	case *EngineContext_V2T:
		return errors.Wrap(t.Valid(), "v2t")
	case *EngineContext_Ppc:
		return errors.Wrap(t.Valid(), "ppc")
	default:
		return errors.Errorf("unknown config type: %T", t)
	}

}

func (e *EventsPull) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return nil
}

func (e *EventsPush) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(e.Events.Valid(), "events")
}

func (e *EngineStream_Pull) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(e.Pull.Valid(), "pull")
}

func (e *EngineStream_Push) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(e.Push.Valid(), "push")
}

func (e *EngineContextEnd) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return nil
}

func (e *EngineStream_End) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(e.End.Valid(), "end")
}

func (e *EngineContextStart) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(e.Context.Valid(), "context")
}

func (e *EngineStream_Start) Valid() error {
	if e == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(e.Start.Valid(), "start")
}

func (l *EngineStream) Valid() error {
	if l == nil {
		return errors.New("is nil")
	}

	switch t := l.Payload.(type) {
	case *EngineStream_Start:
		return errors.Wrap(t.Valid(), "start")
	case *EngineStream_End:
		return errors.Wrap(t.Valid(), "end")
	case *EngineStream_Push:
		return errors.Wrap(t.Valid(), "push")
	case *EngineStream_Pull:
		return errors.Wrap(t.Valid(), "pull")
	default:
		return errors.Errorf("unknown payload type: %T", t)
	}
}

func (i *Lexicon_LexItem) Valid() error {
	if i == nil {
		return errors.New("is nil")
	}

	switch t := i.Item.(type) {
	case *Lexicon_LexItem_User:
		if err := t.Valid(); err != nil {
			return errors.Wrap(err, "user")
		}
	case *Lexicon_LexItem_Main:
		if err := t.Valid(); err != nil {
			return errors.Wrap(err, "main")
		}
	case *Lexicon_LexItem_Noise:
		if err := t.Valid(); err != nil {
			return errors.Wrap(err, "noise")
		}
	default:
		return errors.Errorf("unknown item type: %T", t)
	}

	return nil
}

func (l *Lexicon_UserItem) Valid() error {
	if l == nil {
		return errors.New("is nil")
	}

	if l.Sym == "" {
		return errors.New("sym: blank")
	}

	return nil
}

func (l *Lexicon_MainItem) Valid() error {
	if l == nil {
		return errors.New("is nil")
	}

	return nil
}

func (l *Lexicon_NoiseItem) Valid() error {
	if l == nil {
		return errors.New("is nil")
	}

	return nil
}

func (u *Lexicon_LexItem_User) Valid() error {
	if u == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(u.User.Valid(), "user")
}

func (u *Lexicon_LexItem_Main) Valid() error {
	if u == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(u.Main.Valid(), "main")
}

func (u *Lexicon_LexItem_Noise) Valid() error {
	if u == nil {
		return errors.New("is nil")
	}

	return errors.Wrap(u.Noise.Valid(), "noise")
}

func EndStream() *EngineStream {
	return &EngineStream{
		Payload: &EngineStream_End{
			End: &EngineContextEnd{},
		},
	}
}

func EndStreamWithError(error string) *EngineStream {
	return &EngineStream{
		Payload: &EngineStream_End{
			End: &EngineContextEnd{
				Error: error,
			},
		},
	}
}

func StartStream(ctx *EngineContext) *EngineStream {
	return &EngineStream{Payload: &EngineStream_Start{
		Start: &EngineContextStart{
			Context: ctx,
		},
	}}
}

func IsContextStart(s *EngineStream) bool {
	switch s.Payload.(type) {
	case *EngineStream_Start:
		return true
	default:
		return false

	}
}

func PullStream() *EngineStream {
	return &EngineStream{
		Payload: &EngineStream_Pull{
			Pull: &EventsPull{},
		},
	}
}

func PushStream(evs ...*Event) *EngineStream {
	return &EngineStream{
		Payload: &EngineStream_Push{
			Push: &EventsPush{
				Events: &Events{
					Events:    evs,
					Lookahead: false,
				},
			},
		},
	}
}

func V2TMessageOffsetFn(m *EngineStream) (uint64, bool) {
	if push := m.GetPush(); push != nil {
		if push.Events != nil {
			resOff := uint64(0)
			resHas := false
			for _, e := range push.Events.Events {
				off, has := e.HasOffset()
				if has {
					resOff, resHas = off, has
				}
			}
			return resOff, resHas
		}
	}
	return 0, false
}
