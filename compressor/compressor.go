package compressor

// Compressor is interface, each compressor has Zip and Unzip functions
type Compressor interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
	Type() string
}
