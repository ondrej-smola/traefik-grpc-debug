package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
)

const (
	ContentTypeProtobuf = "application/protobuf"
	ContentTypeJson     = "application/json"
	AcceptHeader        = "Accept"
	ContentTypeHeader   = "Content-Type"
)

var DefaultClient = &http.Client{
	Timeout: time.Second * 10,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	},
}

func WriteError(w http.ResponseWriter, err error, accept string, status int) {
	WriteErrorS(w, err.Error(), accept, status)
}

func WriteErrorS(w http.ResponseWriter, err string, accept string, status int) error {
	switch accept {
	case ContentTypeProtobuf:
		return writeProto(w, &ntx.Error{Error: err}, status)
	case ContentTypeJson:
		fallthrough
	default:
		return WriteJson(w, map[string]string{"error": err}, status)
	}
}

func WriteProtoMsg(w http.ResponseWriter, v proto.Message, contentType string, status int) error {
	switch contentType {
	case ContentTypeProtobuf:
		return writeProto(w, v, status)
	case ContentTypeJson:
		fallthrough
	default:
		return writeJsonPb(w, v, status)
	}
}

func ReadProtoMsg(r io.ReadCloser, v proto.Message, contentType string) error {
	defer r.Close()
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	switch contentType {
	case ContentTypeProtobuf:
		return proto.Unmarshal(body, v)
	case "":
		fallthrough
	case ContentTypeJson:
		return defaultJsonPbUnmarshaller.Unmarshal(bytes.NewReader(body), v)
	default:
		return errors.Errorf("unsupported content type: '%v'", contentType)
	}
}

func ReadJsonMsg(r io.ReadCloser, v proto.Message) error {
	defer r.Close()
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}

var defaultJsonPbMarshaller = jsonpb.Marshaler{}
var defaultJsonPbUnmarshaller = &jsonpb.Unmarshaler{AllowUnknownFields: true}

func writeJsonPb(w http.ResponseWriter, v proto.Message, status int) error {
	head := w.Header()
	head.Add("Content-Type", ContentTypeJson)
	w.WriteHeader(status)
	return errors.Wrap(defaultJsonPbMarshaller.Marshal(w, v), "write")
}

func writeProto(w http.ResponseWriter, v proto.Message, status int) error {
	head := w.Header()
	head.Add("Content-Type", ContentTypeProtobuf)
	w.WriteHeader(status)

	_, err := w.Write(mustProto(v))
	return errors.Wrap(err, "write")
}

func WriteJson(w http.ResponseWriter, v interface{}, status int) error {
	if body, err := json.Marshal(v); err != nil {
		return errors.Wrap(err, "Json marshal")
	} else {
		head := w.Header()
		head.Add("Content-Type", ContentTypeJson)
		w.WriteHeader(status)
		_, err := w.Write(body)
		return errors.Wrap(err, "Response write")
	}
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodOptions {
		WriteError(w, fmt.Errorf("not found: %v %v", r.Method, r.URL.Path), r.Header.Get(AcceptHeader), http.StatusNotFound)
	}
}

func CreatedErrorFromBodyWhenNotExpectedStatusCode(resp *http.Response, code int) error {
	if resp.StatusCode != code {
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return errors.Errorf("unexpected status code %v (expected %v): %v", resp.StatusCode, code, body)
	} else {
		return nil
	}
}

func mustProto(m proto.Message) []byte {
	r, err := proto.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("Must proto: %v", err))
	}
	return r
}

func RemoveDuplicateSlashesFromPath(routePrefix string, route string) string {
	return strings.Replace(routePrefix+route, "//", "/", -1)
}
