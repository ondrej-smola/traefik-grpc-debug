package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/httputil"

	"time"

	"github.com/go-kit/kit/log"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/auth"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/gateway"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/middleware"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/util"
	"gitlab.com/nanotrix/nanogrid/pkg/ntx/v2t/engine"
	"google.golang.org/grpc"
)

type (
	WsToGrpc struct {
		gwFn             gateway.DoFn
		wsTkFn           middleware.WsNtxTokenFn
		jsonMarshaller   jsonpb.Marshaler
		jsonUnmarshaller jsonpb.Unmarshaler
		upgrader         *websocket.Upgrader

		log log.Logger
	}

	wsToGrpcSrc struct {
		jsonMarshaller   jsonpb.Marshaler
		jsonUnmarshaller jsonpb.Unmarshaler

		conn              *websocket.Conn
		ntxToken          *auth.NtxToken
		outputBinaryProto bool
	}
)

const WsMaxWriteDelay = 10 * time.Second
const WsMaxReadLimit = 4 * 1024 * 1024 // 4MB
const WsMaxCloseMessageReasonSizeBytes = 123

func NewWSToGrpc(gwFn gateway.DoFn, wsTkFn middleware.WsNtxTokenFn) *WsToGrpc {
	return &WsToGrpc{
		gwFn:             gwFn,
		wsTkFn:           wsTkFn,
		jsonMarshaller:   jsonpb.Marshaler{},
		jsonUnmarshaller: jsonpb.Unmarshaler{AllowUnknownFields: true},
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		log: log.NewNopLogger(),
	}
}

func isOutputBinaryProto(vals url.Values) bool {
	ok, _ := strconv.ParseBool(vals.Get("binary"))
	return ok
}

func isFlowControlDisabled(vals url.Values) bool {
	ok, _ := strconv.ParseBool(vals.Get("no-flow-control"))
	return ok
}

func (p *WsToGrpc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		httputil.WriteError(w, errors.Wrap(err, "ws upgrade"), r.Header.Get(httputil.AcceptHeader), http.StatusBadRequest)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	conn.SetReadLimit(WsMaxReadLimit)
	query := r.URL.Query()

	src := gateway.Src(&wsToGrpcSrc{
		jsonMarshaller:    p.jsonMarshaller,
		jsonUnmarshaller:  p.jsonUnmarshaller,
		conn:              conn,
		outputBinaryProto: isOutputBinaryProto(query),
	})

	if isFlowControlDisabled(query) {
		src = NewPushOnlySrc(src, ctx)
	}

	ntxTkn, err := p.wsTkFn(query)
	if err != nil {
		conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseUnsupportedData, err.Error()),
			time.Now().Add(WsMaxWriteDelay))
		return
	}

	if ntxTkn.Task == nil {
		conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseUnsupportedData, auth.ErrMissingNtxTokenTask.Error()),
			time.Now().Add(WsMaxWriteDelay))
		return
	}

	reqId := util.NewUUID()

	if err := p.gwFn(reqId, src, func(conn *grpc.ClientConn) (gateway.Dest, error) {
		cl, err := engine.NewEngineServiceClient(conn).StreamingRecognize(ctx)
		return &grpcDest{cl}, err
	}, ntxTkn, ctx); err != nil {
		p.log.Log("ev", "req_err", "id", reqId, "err", err)

		errMsg := err.Error()
		if len(errMsg) > WsMaxCloseMessageReasonSizeBytes {
			errMsg = errMsg[len(errMsg)-WsMaxCloseMessageReasonSizeBytes:]
		}

		if gateway.IsLimitError(err) {
			if err := conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseTryAgainLater, errMsg),
				time.Now().Add(WsMaxWriteDelay)); err != nil {
				p.log.Log("ev", "write_close_msg", "err", err)
			}
		} else {
			if err := conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseInternalServerErr, errMsg),
				time.Now().Add(WsMaxWriteDelay)); err != nil {
				p.log.Log("ev", "write_close_msg", "err", err)
			}
		}
	}

}

func (w *wsToGrpcSrc) SendHeaders(headers map[string]string) error {
	// not supported for WS
	return nil
}

func (w *wsToGrpcSrc) RecvMsg() (*engine.EngineStream, error) {
	for {
		mt, mp, err := w.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				return nil, io.EOF
			} else {
				return nil, err
			}
		}

		msg := &engine.EngineStream{}

		switch mt {
		case websocket.BinaryMessage:
			if err := proto.Unmarshal(mp, msg); err != nil {
				return nil, errors.Wrap(err, "binary msg read")
			} else {
				return msg, msg.Valid()
			}
		case websocket.TextMessage:
			if err := w.jsonUnmarshaller.Unmarshal(bytes.NewReader(mp), msg); err != nil {
				return nil, errors.Wrap(err, "text msg read")
			} else {
				return msg, msg.Valid()
			}
		case websocket.CloseMessage:
			return nil, io.EOF
		default:
			// skip message
		}
	}
}

func (w *wsToGrpcSrc) SendMsg(m *engine.EngineStream) error {
	if w.outputBinaryProto {
		msgBytes, err := proto.Marshal(m)
		if err != nil {
			return err
		}

		if err := w.conn.WriteMessage(websocket.BinaryMessage, msgBytes); err != nil {
			return errors.Wrap(err, "write")
		}
	} else {
		msgString, err := w.jsonMarshaller.MarshalToString(m)
		if err != nil {
			return err
		}

		if err := w.conn.WriteMessage(websocket.TextMessage, []byte(msgString)); err != nil {
			return errors.Wrap(err, "write")
		}
	}

	return nil
}
