package protocol

import (
	"bytes"
	"encoding/binary"
)

// Focus Headers structure looks like:
// +--------------+----------------+----------------+------------+----------+
// | Header count |      Key       |       Value    |    Key     |  Value   |
// +--------------+----------------+----------------+------------+----------+
// |     int16    | int16 + string | int16 + string |           ...         |
// +--------------+----------------+----------------+------------+----------+
type Headers map[string]string

// Marshal will encode request header into a byte slice
func (h Headers) Marshal() []byte {
	var buffer = new(bytes.Buffer)
	var count = int16(len(h))
	if count > 0 {
		err := binary.Write(buffer, binary.BigEndian, count)
		if err != nil {
			return nil
		}
		for k, v := range h {
			WriteString(k, buffer)
			WriteString(v, buffer)
		}
	}
	return buffer.Bytes()
}

// Unmarshal will decode request header into a byte slice
func (h Headers) Unmarshal(data []byte) (err error) {
	if len(data) > 0 {
		buffer := bytes.NewBuffer(data)
		var count int16
		binary.Read(buffer, binary.BigEndian, &count)
		for i := 0; i < int(count); i++ {
			k := ReadString(buffer)
			v := ReadString(buffer)
			h[k] = v
		}
	}
	return
}

func WriteString(data string, buffer *bytes.Buffer) {
	if len(data) == 0 {
		binary.Write(buffer, binary.BigEndian, int16(0))
	} else {
		dbs := []byte(data)
		var len = int16(len(dbs))
		binary.Write(buffer, binary.BigEndian, len)
		binary.Write(buffer, binary.BigEndian, dbs)
	}
}

func ReadString(buffer *bytes.Buffer) string {
	var len int16
	binary.Read(buffer, binary.BigEndian, &len)
	if len > 0 {
		dbs := make([]byte, len)
		binary.Read(buffer, binary.BigEndian, dbs)
		return string(dbs)
	} else {
		return ""
	}
}
