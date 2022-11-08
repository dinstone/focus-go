package codec

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/dinstone/focus-go/protocol"
	"io"
	"net/rpc"
	"strings"
	"sync"

	"github.com/dinstone/focus-go/compressor"
	"github.com/dinstone/focus-go/serializer"
)

type clientCodec struct {
	r io.Reader
	w io.Writer
	c io.Closer

	compressor compressor.Compressor // rpc compress type(raw,gzip,snappy,zlib)
	serializer serializer.Serializer // json,protobuf
	response   protocol.Message
	mutex      sync.Mutex // protect pending map
	pending    map[uint64]string
}

// NewClientCodec Create a new client codec
func NewClientCodec(conn io.ReadWriteCloser,
	compressor compressor.Compressor, serializer serializer.Serializer) rpc.ClientCodec {
	if compressor == nil {
		//compressor = compressor.RawCompressor{}
	}
	return &clientCodec{
		r:          bufio.NewReader(conn),
		w:          bufio.NewWriter(conn),
		c:          conn,
		compressor: compressor,
		serializer: serializer,
		pending:    make(map[uint64]string),
	}
}

// WriteRequest Write the rpc requestHeader header and body to the io stream
func (c *clientCodec) WriteRequest(request *rpc.Request, param interface{}) error {
	c.mutex.Lock()
	c.pending[request.Seq] = request.ServiceMethod
	c.mutex.Unlock()

	reqBody, err := c.serializer.Encode(param)
	if err != nil {
		return err
	}
	compressedReqBody, err := c.compressor.Encode(reqBody)
	if err != nil {
		return err
	}

	dot := strings.LastIndex(request.ServiceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc: service/method request ill-formed: " + request.ServiceMethod)
		return err
	}
	serviceName := request.ServiceMethod[:dot]
	methodName := request.ServiceMethod[dot+1:]

	// version
	err = binary.Write(c.w, binary.BigEndian, int8(1))
	if err != nil {
		return err
	}
	if methodName == "Heartbeat" {
		// protocol type: heartbeat
		err = binary.Write(c.w, binary.BigEndian, int8(0))
		if err != nil {
			return err
		}
		err = binary.Write(c.w, binary.BigEndian, int8(1))
		if err != nil {
			return err
		}
	} else {
		// protocol type: request
		err = binary.Write(c.w, binary.BigEndian, int8(1))
		if err != nil {
			return err
		}
		// protocol timeout
		err = binary.Write(c.w, binary.BigEndian, int16(3000))
		if err != nil {
			return err
		}
	}

	// protocol seq
	err = binary.Write(c.w, binary.BigEndian, int32(request.Seq))
	if err != nil {
		return err
	}

	headers := make(protocol.Headers, 2)
	headers["call.service"] = serviceName
	headers["call.method"] = methodName
	headers["serializer.type"] = c.serializer.Type()
	headers["compressor.type"] = c.compressor.Type()
	headersData := headers.Marshal()
	// headers length
	x := len(headersData)
	err = binary.Write(c.w, binary.BigEndian, int32(x))
	if err != nil {
		return err
	}
	if x > 0 {
		err = binary.Write(c.w, binary.BigEndian, headersData)
		if err != nil {
			return err
		}
	}

	y := len(compressedReqBody)
	err = binary.Write(c.w, binary.BigEndian, int32(y))
	if err != nil {
		return err
	}
	if y > 0 {
		err = binary.Write(c.w, binary.BigEndian, compressedReqBody)
		if err != nil {
			return err
		}
	}

	c.w.(*bufio.Writer).Flush()
	return nil
}

// ReadResponseHeader read the rpc responseHeader header from the io stream
func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	var version int8
	err := binary.Read(c.r, binary.BigEndian, &version)
	if err != nil {
		return err
	}
	var msgType int8
	err = binary.Read(c.r, binary.BigEndian, &msgType)
	if err != nil {
		return err
	}
	var flag int16
	err = binary.Read(c.r, binary.BigEndian, &flag)
	if err != nil {
		return err
	}
	var msgId int32
	err = binary.Read(c.r, binary.BigEndian, &msgId)
	if err != nil {
		return err
	}
	// read headers
	var hl int32
	err = binary.Read(c.r, binary.BigEndian, &hl)
	if err != nil {
		return err
	}
	headers := make(protocol.Headers, 2)
	if hl > 0 {
		buf := make([]byte, hl)
		err = binary.Read(c.r, binary.BigEndian, buf)
		if err != nil {
			return err
		}
		headers.Unmarshal(buf)
	}

	var bl int32
	err = binary.Read(c.r, binary.BigEndian, &bl)
	if err != nil {
		return err
	}
	responseBody := make([]byte, bl)
	err = binary.Read(c.r, binary.BigEndian, responseBody)
	if err != nil {
		return err
	}

	if headers["compressor.type"] != "" {
		responseBody, err = c.compressor.Decode(responseBody)
		if err != nil {
			return err
		}
	}

	// response protobuf
	if flag == 1 {
		buffer := bytes.NewBuffer(responseBody)
		var code int32
		binary.Read(buffer, binary.BigEndian, &code)
		es := protocol.ReadString(buffer)
		r.Error = es
	}

	c.response.Version = version
	c.response.MsgType = msgType
	c.response.Flag = flag
	c.response.MsgId = msgId
	c.response.Headers = headers
	c.response.Content = responseBody

	c.mutex.Lock()
	r.Seq = uint64(msgId)
	// r.Error = c.responseHeader.Error
	r.ServiceMethod = c.pending[r.Seq]
	delete(c.pending, r.Seq)
	c.mutex.Unlock()
	return nil
}

// ReadResponseBody read the rpc responseHeader body from the io stream
func (c *clientCodec) ReadResponseBody(param interface{}) error {
	if param == nil {
		return nil
	}

	// heartbeat protobuf
	if c.response.MsgType == 0 {
		return nil
	}

	return c.serializer.Decode(c.response.Content, param)
}

func (c *clientCodec) Close() error {
	return c.c.Close()
}
