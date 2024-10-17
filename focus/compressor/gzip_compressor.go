package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
)

var Gzip = GzipCompressor{}

// GzipCompressor implements the Compressor interface
type GzipCompressor struct {
}

// Zip .
func (_ GzipCompressor) Encode(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := gzip.NewWriter(buf)
	defer w.Close()
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

// Unzip .
func (_ GzipCompressor) Decode(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err = ioutil.ReadAll(r)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return data, nil
}

func (GzipCompressor) Type() string {
	return "gzip"
}
