package serializer

type Serializer interface {
	Encode(message interface{}) ([]byte, error)
	Decode(data []byte, message interface{}) error
	Type() string
}
