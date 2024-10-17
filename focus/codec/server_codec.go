package codec

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"net/rpc"
	"sync"

	"github.com/dinstone/focus-go/focus/protocol"

	"github.com/dinstone/focus-go/focus/compressor"
	"github.com/dinstone/focus-go/focus/serializer"
)

type reqCtx struct {
	requestID  uint64
	msgType    int8
	compressor compressor.Compressor
}

type serverCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	serializer serializer.Serializer
	compressor compressor.Compressor
	mutex      sync.Mutex // protects seq, pending
	seq        uint64
	pending    map[uint64]*reqCtx
}

// NewServerCodec Create a new server codec
func NewServerCodec(conn io.ReadWriteCloser, serializer serializer.Serializer, compressor compressor.Compressor) rpc.ServerCodec {
	return &serverCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		serializer: serializer,
		compressor: compressor,
		pending:    make(map[uint64]*reqCtx),
	}
}

// ReadRequestHeader read the rpc requestHeader header from the io stream
func (s *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	var version int8
	err := binary.Read(s.r, binary.BigEndian, &version)
	if err != nil {
		return err
	}
	var msgType int8
	err = binary.Read(s.r, binary.BigEndian, &msgType)
	if err != nil {
		return err
	}
	var flag int16
	err = binary.Read(s.r, binary.BigEndian, &flag)
	if err != nil {
		return err
	}
	var msgId int32
	err = binary.Read(s.r, binary.BigEndian, &msgId)
	if err != nil {
		return err
	}
	// read headers
	var hl int32
	err = binary.Read(s.r, binary.BigEndian, &hl)
	if err != nil {
		return err
	}
	headers := make(protocol.Headers, 2)
	if hl > 0 {
		buf := make([]byte, hl)
		err = binary.Read(s.r, binary.BigEndian, buf)
		if err != nil {
			return err
		}
		headers.Unmarshal(buf)
	}

	s.mutex.Lock()
	s.seq = uint64(msgId)
	s.pending[s.seq] = &reqCtx{s.seq, msgType, s.compressor}
	r.ServiceMethod = headers["call.service"] + "." + headers["call.method"]
	r.Seq = s.seq
	s.mutex.Unlock()
	return nil
}

// ReadRequestBody read the rpc requestHeader body from the io stream
func (s *serverCodec) ReadRequestBody(param interface{}) error {
	if param == nil {
		var bl int32
		binary.Read(s.r, binary.BigEndian, &bl)
		err := binary.Read(s.r, binary.BigEndian, make([]byte, bl))
		if err != nil {
			return err
		}
		return nil
	}

	var bl int32
	err := binary.Read(s.r, binary.BigEndian, &bl)
	if err != nil {
		return err
	}
	reqBody := make([]byte, bl)
	err = binary.Read(s.r, binary.BigEndian, reqBody)
	if err != nil {
		return err
	}

	req, err := s.compressor.Decode(reqBody)
	if err != nil {
		return err
	}

	return s.serializer.Decode(req, param)
}

// WriteResponse Write the rpc responseHeader header and body to the io stream
func (s *serverCodec) WriteResponse(r *rpc.Response, param interface{}) error {
	s.mutex.Lock()
	reqCtx, _ := s.pending[r.Seq]
	delete(s.pending, r.Seq)
	s.mutex.Unlock()

	// heartbeat protobuf
	if reqCtx.msgType == 0 {
		// version
		binary.Write(s.w, binary.BigEndian, int8(1))
		// protobuf type
		binary.Write(s.w, binary.BigEndian, int8(0))
		// protobuf status
		binary.Write(s.w, binary.BigEndian, int16(0))
		// protobuf seq
		binary.Write(s.w, binary.BigEndian, int32(r.Seq))
		// headers
		binary.Write(s.w, binary.BigEndian, int32(0))
		// content
		err := binary.Write(s.w, binary.BigEndian, int32(0))
		if err != nil {
			return err
		}
		return nil
	}

	// response protobuf
	// version
	err := binary.Write(s.w, binary.BigEndian, int8(1))
	if err != nil {
		return err
	}
	// protobuf type
	binary.Write(s.w, binary.BigEndian, int8(2))
	if err != nil {
		return err
	}
	if r.Error != "" {
		// protobuf status
		binary.Write(s.w, binary.BigEndian, int16(1))
		if err != nil {
			return err
		}
	} else {
		// protobuf status
		binary.Write(s.w, binary.BigEndian, int16(0))
		if err != nil {
			return err
		}
	}

	// protobuf seq
	binary.Write(s.w, binary.BigEndian, int32(r.Seq))
	if err != nil {
		return err
	}

	headers := make(protocol.Headers, 2)
	var respBody []byte
	//var err error
	if r.Error != "" {
		var buffer = new(bytes.Buffer)
		err := binary.Write(buffer, binary.BigEndian, int32(999))
		if err != nil {
			return err
		}
		protocol.WriteString(r.Error, buffer)
		respBody = buffer.Bytes()
	} else {
		respBody, err = s.serializer.Encode(param)
		if err != nil {
			return err
		}
		respBody, err = s.compressor.Encode(respBody)
		if err != nil {
			return err
		}
		headers["serializer.type"] = s.serializer.Type()
		headers["compressor.type"] = s.compressor.Type()
	}

	// headers length
	hdata := headers.Marshal()
	x := len(hdata)
	if x == 0 {
		binary.Write(s.w, binary.BigEndian, int32(x))
	} else {
		binary.Write(s.w, binary.BigEndian, int32(x))
		// write(s.w, hdata)
		binary.Write(s.w, binary.BigEndian, hdata)
	}

	y := len(respBody)
	if y == 0 {
		binary.Write(s.w, binary.BigEndian, int32(y))
	} else {
		binary.Write(s.w, binary.BigEndian, int32(y))
		// write(s.w, respBody)
		binary.Write(s.w, binary.BigEndian, respBody)
	}

	s.w.(*bufio.Writer).Flush()
	return nil
}

func (s *serverCodec) Close() error {
	return s.c.Close()
}
