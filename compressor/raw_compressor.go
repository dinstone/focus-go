package compressor

var Raw = RawCompressor{}

// RawCompressor implements the Compressor interface
type RawCompressor struct {
}

// Zip .
func (_ RawCompressor) Encode(data []byte) ([]byte, error) {
	return data, nil
}

// Unzip .
func (_ RawCompressor) Decode(data []byte) ([]byte, error) {
	return data, nil
}

func (RawCompressor) Type() string {
	return "raw"
}
