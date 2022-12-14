package options

import (
	"github.com/dinstone/focus-go/compressor"
	"github.com/dinstone/focus-go/serializer"
)

type Options struct {
	Compressor compressor.Compressor
	Serializer serializer.Serializer
}

type ServerOptions struct {
	Options
	Host string
	Port int
}

type Option func(o *Options)

// WithCompressor set client compression format
func WithCompressor(c compressor.Compressor) Option {
	return func(o *Options) {
		if c != nil {
			o.Compressor = c
		}
	}
}

// WithSerializer set client serializer
func WithSerializer(serializer serializer.Serializer) Option {
	return func(o *Options) {
		o.Serializer = serializer
	}
}
