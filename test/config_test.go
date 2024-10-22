package test

import (
	"fmt"
	"testing"

	"github.com/dinstone/focus-go/focus/compressor"
	"github.com/dinstone/focus-go/focus/serializer"
)

func TestConfig(t *testing.T) {
	config := NewConfig()
	config.ShowConfig()

	config.SetCompressor(compressor.Gzip)
	config.SetSerializer(serializer.Protobuf)
	config.ShowConfig()

	fmt.Println(config.serializer.Type())

	c := Config{compressor: compressor.Raw, serializer: serializer.Json}
	show(c)
	update(c)
	show(c)
	handle(&c)
	show(c)
}

func update(c Config) {
	c.compressor = compressor.Gzip
	c.serializer = serializer.Protobuf
}

func handle(c *Config) {
	c.compressor = compressor.Gzip
	c.serializer = serializer.Protobuf
}

func show(c Config) {
	fmt.Printf("compressor = %s; serializer = %s \n", c.compressor.Type(), c.serializer.Type())
}
