package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"gitlab.com/nanotrix/nanogrid/pkg/ntx/httputil"

	"bytes"
	"encoding/base64"
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
)

type readerNoopCloser struct {
	from io.Reader
}

func (r *readerNoopCloser) Read(p []byte) (int, error) {
	return r.from.Read(p)
}
func (r *readerNoopCloser) Close() error {
	return nil
}

type BufferCloser struct {
	*bytes.Buffer
}

func (cb *BufferCloser) Close() error {
	return nil
}

func ToBase64JsonInputString(in interface{}) (string, error) {
	body, err := json.Marshal(in)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("base64:%v", base64.StdEncoding.EncodeToString(body)), nil
}

type InputType int

const (
	InputType_Unknown InputType = iota
	InputType_Base64
	InputType_Env
	InputType_Base64Env
	InputType_File
	InputType_Http
	InputType_Https
	InputType_String
	InputType_Stdin
)

const InputTypeFilePrefix = "file:"
const InputTypeBase64Prefix = "base64:"
const InputTypeEnvBase64Prefix = "envbase64:"
const InputTypeEnvPrefix = "env:"
const InputTypeHttpPrefix = "http:"
const InputTypeHttpsPrefix = "https:"
const InputTypeStringPrefix = "string:"

func GetInputType(input string) InputType {
	if input == "-" {
		return InputType_Stdin
	} else if strings.HasPrefix(input, InputTypeBase64Prefix) {
		return InputType_Base64
	} else if strings.HasPrefix(input, InputTypeEnvBase64Prefix) {
		return InputType_Base64Env
	} else if strings.HasPrefix(input, InputTypeEnvPrefix) {
		return InputType_Env
	} else if strings.HasPrefix(input, InputTypeFilePrefix) {
		return InputType_File
	} else if strings.HasPrefix(input, InputTypeHttpsPrefix) {
		return InputType_Https
	} else if strings.HasPrefix(input, InputTypeHttpPrefix) {
		return InputType_Http
	} else if strings.HasPrefix(input, InputTypeStringPrefix) {
		return InputType_String
	} else {
		return InputType_String
	}
}

func ReadFromInputString(input string) (io.ReadCloser, error) {
	switch GetInputType(input) {
	case InputType_Stdin:
		return &readerNoopCloser{os.Stdin}, nil
	case InputType_Base64:
		base := strings.TrimPrefix(input, InputTypeBase64Prefix)
		if body, err := base64.StdEncoding.DecodeString(base); err != nil {
			return nil, err
		} else {
			return &BufferCloser{bytes.NewBuffer(body)}, nil
		}
	case InputType_Base64Env:
		e := strings.TrimPrefix(input, InputTypeEnvBase64Prefix)
		base := os.Getenv(e)
		if base == "" {
			return nil, errors.Errorf("Env not set: '%v'", e)
		}

		if body, err := base64.StdEncoding.DecodeString(base); err != nil {
			return nil, err
		} else {
			return &BufferCloser{bytes.NewBuffer(body)}, nil
		}
	case InputType_Env:
		e := strings.TrimPrefix(input, InputTypeEnvPrefix)
		base := os.Getenv(e)
		if base == "" {
			return nil, errors.Errorf("Env not set: '%v'", e)
		}
		return &BufferCloser{bytes.NewBuffer([]byte(base))}, nil
	case InputType_File:
		path := strings.TrimPrefix(input, InputTypeFilePrefix)
		reader, err := os.Open(path)
		return reader, errors.Wrapf(err, "Failed to open '%v'", path)
	case InputType_Http:
		fallthrough
	case InputType_Https:
		reader, err := httputil.DefaultClient.Get(input)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to http get '%v'", input)
		} else {
			return reader.Body, nil
		}
	case InputType_String:
		base := strings.TrimPrefix(input, InputTypeStringPrefix)
		return &BufferCloser{bytes.NewBuffer([]byte(base))}, nil
	default:
		return nil, errors.Errorf("Unsupported input string '%v'", input)
	}
}

func ReadBytesFromInputString(input string) ([]byte, error) {
	r, err := ReadFromInputString(input)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

func ReadStringFromInputString(input string) (string, error) {
	r, err := ReadBytesFromInputString(input)
	if err != nil {
		return "", err
	}
	return string(r), nil
}

func ReadAllFromInputStringToJson(input string, to interface{}) error {
	r, err := ReadFromInputString(input)
	if err != nil {
		return err
	}
	defer r.Close()
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, to)
}

func FileToJson(input string, to interface{}) error {
	r, err := ReadFromInputString(InputTypeFilePrefix + input)
	if err != nil {
		return err
	}
	defer r.Close()
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, to)
}

var defaultJsonPbUnmarshaller = &jsonpb.Unmarshaler{AllowUnknownFields: true}
var defaultJSONPBMarshaller = jsonpb.Marshaler{}

func FileToJsonPb(input string, to proto.Message) error {
	r, err := ReadFromInputString(InputTypeFilePrefix + input)
	if err != nil {
		return err
	}
	defer r.Close()
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return defaultJsonPbUnmarshaller.Unmarshal(bytes.NewReader(body), to)
}

func ReadAllFromInputStringToJsonPb(input string, to proto.Message) error {
	r, err := ReadFromInputString(input)
	if err != nil {
		return err
	}
	defer r.Close()
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return defaultJsonPbUnmarshaller.Unmarshal(bytes.NewReader(body), to)
}

type writerNoopCloser struct {
	to io.Writer
}

func (r *writerNoopCloser) Write(p []byte) (int, error) {
	return r.to.Write(p)
}

func (r *writerNoopCloser) Close() error {
	return nil
}

func WriteToOutputString(output string) (io.WriteCloser, error) {
	if output == "-" {
		return &writerNoopCloser{os.Stdout}, nil
	} else if strings.HasPrefix(output, "file:") {
		path := strings.TrimPrefix(output, "file:")
		writer, err := os.Open(path)
		return writer, errors.Wrapf(err, "Failed to open '%v'", path)
	} else {
		return nil, errors.Errorf("Unsupported output '%v'", output)
	}
}

func ReadCloserToBytes(r io.ReadCloser) ([]byte, error) {
	defer r.Close()
	return ioutil.ReadAll(r)
}

func ReadCloserToString(r io.ReadCloser) (string, error) {
	if b, err := ReadCloserToBytes(r); err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

func ReadCloserToJson(r io.ReadCloser, to interface{}) error {
	body, err := ReadCloserToBytes(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &to)
}

func ReadCloserToJsonPb(r io.ReadCloser, to proto.Message) error {
	body, err := ReadCloserToBytes(r)
	if err != nil {
		return err
	}

	return defaultJsonPbUnmarshaller.Unmarshal(bytes.NewReader(body), to)
}

func ParseJsonPb(body []byte, to proto.Message) error {
	return defaultJsonPbUnmarshaller.Unmarshal(bytes.NewReader(body), to)
}

func MustJsonPb(m proto.Message) []byte {
	r, err := defaultJSONPBMarshaller.MarshalToString(m)
	if err != nil {
		panic(fmt.Sprintf("Must jsonpb: %v", err))
	}
	return []byte(r)
}

func MustProto(m proto.Message) []byte {
	r, err := proto.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("Must proto: %v", err))
	}
	return r
}

func MustJson(m interface{}) []byte {
	if body, err := json.Marshal(m); err != nil {
		panic(fmt.Sprintf("must json: %v", err))
	} else {
		return body
	}
}

func CloneBytes(in []byte) []byte {
	cl := make([]byte, len(in))
	copy(cl, in)
	return cl
}
