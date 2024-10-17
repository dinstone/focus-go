package compressor

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/golang/snappy"
)

var Snappy = SnappyCompressor{}

// SnappyCompressor implements the Compressor interface
type SnappyCompressor struct {
}

// Zip .
func (_ SnappyCompressor) Encode(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := snappy.NewBufferedWriter(buf)
	defer func() {
		w.Close()
	}()
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
func (_ SnappyCompressor) Decode(data []byte) ([]byte, error) {
	r := snappy.NewReader(bytes.NewBuffer(data))
	data, err := ioutil.ReadAll(r)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return data, nil
}

func (SnappyCompressor) Type() string {
	return "snappy"
}
