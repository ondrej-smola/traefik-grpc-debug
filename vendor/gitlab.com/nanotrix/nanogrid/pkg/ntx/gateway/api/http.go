package api

import (
	"context"
	"io"
	"net/http"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/httputil"

	"strings"

	"github.com/go-kit/kit/log"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/gateway"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/v2t/engine"
	"google.golang.org/grpc"
)

type (
	HttpToGrpc struct {
		gwFn gateway.DoFn

		jsonMarshaller   jsonpb.Marshaler
		jsonUnmarshaller jsonpb.Unmarshaler
		labelParser      engine.LabelToContextParser
		bufferSize       int

		log log.Logger
	}

	httpToGRPCSrc struct {
		innerChan      chan *engine.EngineStream
		jsonMarshaller jsonpb.Marshaler
		buffer         []byte

		ctx         context.Context
		closeReader context.CancelFunc
		readFrom    io.Reader
		writeTo     http.ResponseWriter
	}

	HttpToGrpcOpt func(h *HttpToGrpc)
)

const MultipartMaxInMemoryBytes = 2 << 20 // 2 MB
const FormKeyLexicon = "lexicon"
const FormKeyFile = "file"
const FormKeyChannel = "channel"
const AudioChannelLeft = "left"
const AudioChannelRight = "right"
const AudioChannelDownmix = "downmix"

func WithHttpBufferSize(size int) HttpToGrpcOpt {
	return func(g *HttpToGrpc) {
		g.bufferSize = size
	}
}

func WithHttpLogger(log log.Logger) HttpToGrpcOpt {
	return func(g *HttpToGrpc) {
		g.log = log
	}
}

func NewHttpToGrpc(gwFn gateway.DoFn, opts ...HttpToGrpcOpt) *HttpToGrpc {
	p := &HttpToGrpc{
		gwFn: gwFn,

		jsonMarshaller:   jsonpb.Marshaler{EmitDefaults: true},
		jsonUnmarshaller: jsonpb.Unmarshaler{AllowUnknownFields: true},
		labelParser:      engine.NewLabelToContextParser(),
		bufferSize:       4096,
		log:              log.NewNopLogger(),
	}

	for _, o := range opts {
		o(p)
	}

	return p
}

func (p *HttpToGrpc) setChannel(r *http.Request, ctx *engine.EngineContext) error {
	switch ch := r.FormValue(FormKeyChannel); ch {
	case "":
		ctx.AudioChannel = engine.EngineContext_AUDIO_CHANNEL_DOWNMIX
	case AudioChannelDownmix:
		ctx.AudioChannel = engine.EngineContext_AUDIO_CHANNEL_DOWNMIX
	case AudioChannelRight:
		ctx.AudioChannel = engine.EngineContext_AUDIO_CHANNEL_RIGHT
	case AudioChannelLeft:
		ctx.AudioChannel = engine.EngineContext_AUDIO_CHANNEL_LEFT
	default:
		return errors.Errorf("invalid channel '%v': use one of [%v,%v,%v]",
			ch, AudioChannelDownmix, AudioChannelRight, AudioChannelLeft)
	}

	return nil
}

func (p *HttpToGrpc) setLexicon(r *http.Request, ctx *engine.EngineContext) error {
	v2t := ctx.GetV2T()
	if v2t == nil {
		return nil
	}
	lex := &engine.Lexicon{}

	lexiconStr := r.FormValue(FormKeyLexicon)
	var lexReader io.Reader

	if lexiconStr == "" {
		if file, _, err := r.FormFile(FormKeyLexicon); err != nil {
			if err != http.ErrMissingFile {
				return errors.Wrap(err, FormKeyLexicon)
			} else {
				return nil
			}
		} else {
			defer file.Close()
			lexReader = file
		}
	} else {
		lexReader = strings.NewReader(lexiconStr)
	}

	if err := p.jsonUnmarshaller.Unmarshal(lexReader, lex); err != nil {
		return errors.Wrap(err, "json unmarshal")
	} else if err := lex.Valid(); err != nil {
		return errors.Wrap(err, "lexicon")
	} else {
		v2t.WithLexicon = lex
		return nil
	}
}

func (p *HttpToGrpc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(MultipartMaxInMemoryBytes); err != nil {
		httputil.WriteError(w, errors.Wrap(err, "parse form"), r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
		return
	}

	defer r.MultipartForm.RemoveAll()

	if tkn, ok := auth.GetNtxTokenFromContext(r.Context()); !ok {
		httputil.WriteError(w, auth.ErrMissingNtxToken, r.Header.Get(httputil.AcceptHeader), http.StatusUnauthorized)
	} else if tkn.Task == nil {
		httputil.WriteError(w, auth.ErrMissingNtxTokenTask, r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
	} else if v2tCtx, err := p.labelParser.Parse(tkn.Task.Label); err != nil {
		httputil.WriteError(w, errors.Errorf("create context from label %v", tkn.Task.Label), r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
	} else if file, _, err := r.FormFile(FormKeyFile); err != nil {
		httputil.WriteError(w, errors.Wrap(err, "form 'file'"), r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
	} else {
		defer file.Close()

		if err := p.setLexicon(r, v2tCtx); err != nil {
			httputil.WriteError(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
		} else if err := p.setChannel(r, v2tCtx); err != nil {
			httputil.WriteError(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
		} else {

			ctx, closeReader := context.WithCancel(context.Background())
			defer closeReader()
			reqId := util.NewUUID()

			innerChan := make(chan *engine.EngineStream, 4)
			innerChan <- engine.StartStream(v2tCtx)

			src := &httpToGRPCSrc{
				innerChan:   innerChan,
				ctx:         ctx,
				closeReader: closeReader,
				buffer:      make([]byte, p.bufferSize),
				readFrom:    file,
				writeTo:     w,
			}

			if err := p.gwFn(reqId, src, func(conn *grpc.ClientConn) (gateway.Dest, error) {
				cl, err := engine.NewEngineServiceClient(conn).StreamingRecognize(ctx)
				return &grpcDest{cl}, err
			}, tkn, r.Context()); err != nil {
				if gateway.IsLimitError(err) {
					httputil.WriteError(w, err, r.Header.Get(httputil.AcceptHeader), http.StatusTooManyRequests)
				} else {
					src.write(engine.EndStreamWithError(err.Error()))
					p.log.Log("ev", "req_err", "id", reqId, "err", err)
				}
			} else {
				src.write(engine.EndStream())
			}
		}
	}
}

func (w *httpToGRPCSrc) SendHeaders(headers map[string]string) error {
	wh := w.writeTo.Header()

	for k, v := range headers {
		wh.Set(k, v)
	}

	return nil
}

func (w *httpToGRPCSrc) RecvMsg() (*engine.EngineStream, error) {
	select {
	case <-w.ctx.Done():
		// give innerChan priority in case last item is written together with context cancel
		select {
		case eng := <-w.innerChan:
			return eng, nil
		default:
			return nil, io.EOF
		}
	case eng := <-w.innerChan:
		return eng, nil
	}
}

func (w *httpToGRPCSrc) write(m *engine.EngineStream) error {
	msgString, err := w.jsonMarshaller.MarshalToString(m)
	if err != nil {
		return errors.Wrap(err, "jsonpb")
	}

	if _, err = w.writeTo.Write([]byte(msgString + "\n")); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func (w *httpToGRPCSrc) SendMsg(m *engine.EngineStream) error {
	write := true

	switch m.Payload.(type) {
	case *engine.EngineStream_Start:
		w.innerChan <- engine.PullStream()
	case *engine.EngineStream_Push:
		w.innerChan <- engine.PullStream()
	case *engine.EngineStream_Pull:
		write = false
		n, err := w.readFrom.Read(w.buffer)
		if n > 0 {
			w.innerChan <- engine.PushStream(engine.AudioEvent(w.buffer[:n]))
		} else if err == io.EOF {
			w.innerChan <- engine.EndStream()
		} else {
			w.closeReader()
			return err
		}
	case *engine.EngineStream_End:
		w.closeReader()
		write = false
	}

	if write {
		return w.write(m)
	} else {
		return nil
	}
}
