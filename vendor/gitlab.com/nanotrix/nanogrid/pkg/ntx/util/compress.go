package util

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

func GzipEncode(msg []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(msg); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GzipDecode(msg []byte) ([]byte, error) {
	read, err := gzip.NewReader(bytes.NewReader(msg))
	if err != nil {
		return nil, err
	}
	defer read.Close()
	return ioutil.ReadAll(read)
}
