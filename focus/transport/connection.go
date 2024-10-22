package transport

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"

	"github.com/dinstone/focus-go/focus/protocol"
)

type Connection struct {
	reader io.Reader
	writer io.Writer
	closer io.Closer
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{reader: bufio.NewReader(conn), writer: bufio.NewWriter(conn), closer: conn}
}

func (c *Connection) ReadMessage() (*protocol.Message, error) {
	var version int8
	err := binary.Read(c.reader, binary.BigEndian, &version)
	if err != nil {
		return nil, err
	}
	var msgType int8
	err = binary.Read(c.reader, binary.BigEndian, &msgType)
	if err != nil {
		return nil, err
	}
	var flag int16
	err = binary.Read(c.reader, binary.BigEndian, &flag)
	if err != nil {
		return nil, err
	}
	var msgId int32
	err = binary.Read(c.reader, binary.BigEndian, &msgId)
	if err != nil {
		return nil, err
	}
	// read headers
	var hl int32
	err = binary.Read(c.reader, binary.BigEndian, &hl)
	if err != nil {
		return nil, err
	}
	headers := make(protocol.Headers, 2)
	if hl > 0 {
		buf := make([]byte, hl)
		err = binary.Read(c.reader, binary.BigEndian, buf)
		if err != nil {
			return nil, err
		}
		headers.Unmarshal(buf)
	}

	var cl int32
	err = binary.Read(c.reader, binary.BigEndian, &cl)
	if err != nil {
		return nil, err
	}
	content := make([]byte, cl)
	err = binary.Read(c.reader, binary.BigEndian, content)
	if err != nil {
		return nil, err
	}

	message := new(protocol.Message)
	message.Version = version
	message.MsgType = msgType
	message.Flag = flag
	message.MsgId = msgId
	message.Headers = headers
	message.Content = content

	return message, nil
}

func (c *Connection) WriteMessage(message *protocol.Message) error {

	// version
	binary.Write(c.writer, binary.BigEndian, message.Version)
	//  type
	binary.Write(c.writer, binary.BigEndian, message.MsgType)
	//  status
	binary.Write(c.writer, binary.BigEndian, message.Flag)
	//  Sequence
	binary.Write(c.writer, binary.BigEndian, message.MsgId)

	// headers length
	hdata := message.Headers.Marshal()
	x := len(hdata)
	if x == 0 {
		binary.Write(c.writer, binary.BigEndian, int32(x))
	} else {
		binary.Write(c.writer, binary.BigEndian, int32(x))
		binary.Write(c.writer, binary.BigEndian, hdata)
	}

	y := len(message.Content)
	if y == 0 {
		binary.Write(c.writer, binary.BigEndian, int32(y))
	} else {
		binary.Write(c.writer, binary.BigEndian, int32(y))
		binary.Write(c.writer, binary.BigEndian, message.Content)
	}

	c.writer.(*bufio.Writer).Flush()

	return nil
}

func (c *Connection) Close() {
	c.closer.Close()
}
