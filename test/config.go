package test

import (
	"fmt"

	"github.com/dinstone/focus-go/focus/compressor"
	"github.com/dinstone/focus-go/focus/serializer"
)

type Config struct {
	compressor compressor.Compressor
	serializer serializer.Serializer
}

func NewConfig() *Config {
	return &Config{compressor: compressor.Raw, serializer: serializer.Json}
}

func (c *Config) SetCompressor(compressor compressor.Compressor) {
	c.compressor = compressor
}

func (c *Config) SetSerializer(serializer serializer.Serializer) {
	c.serializer = serializer
}

func (c *Config) ShowConfig() {
	fmt.Printf("s=%s,c=%s \n", c.serializer.Type(), c.compressor.Type())
}
