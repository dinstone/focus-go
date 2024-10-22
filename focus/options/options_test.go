package options

import (
	"fmt"
	"testing"

	"github.com/dinstone/focus-go/focus/compressor"
	"github.com/dinstone/focus-go/focus/serializer"
)

func TestOptions(t *testing.T) {
	so := ServerOptions{
		Options: Options{
			serializer: serializer.Protobuf,
			compressor: compressor.Raw,
		},
		Address: ":8090",
	}
	fmt.Println(so.SetCompressor(compressor.Gzip))
	fmt.Println("server options: ", so)
}
