package serializer

import "encoding/json"

var Json = JsonSerializer{}

// Json .serializer
type JsonSerializer struct{}

// Marshal .
func (_ JsonSerializer) Encode(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}

// Unmarshal .
func (_ JsonSerializer) Decode(data []byte, message interface{}) error {
	return json.Unmarshal(data, message)
}

func (JsonSerializer) Type() string {
	return "json"
}
