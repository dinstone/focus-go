package compressor

import (
	"bytes"
	"compress/zlib"
	"io"
	"io/ioutil"
)

var Zlib = ZlibCompressor{}

// ZlibCompressor implements the Compressor interface
type ZlibCompressor struct {
}

// Zip .
func (_ ZlibCompressor) Encode(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := zlib.NewWriter(buf)
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
func (_ ZlibCompressor) Decode(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewBuffer(data))
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

func (ZlibCompressor) Type() string {
	return "zlib"
}
