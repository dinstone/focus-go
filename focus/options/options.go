package options

import (
	"github.com/dinstone/focus-go/focus/compressor"
	"github.com/dinstone/focus-go/focus/serializer"
)

type Options struct {
	compressor compressor.Compressor
	serializer serializer.Serializer
}

func (o *Options) SetCompressor(c compressor.Compressor) *Options {
	o.compressor = c
	return o
}

func (o *Options) GetCompressor() compressor.Compressor {
	return o.compressor
}

func (o *Options) SetSerializer(s serializer.Serializer) *Options {
	o.serializer = s
	return o
}

func (o *Options) GetSerializer() serializer.Serializer {
	return o.serializer
}

type ClientOptions struct {
	Options
	Address string
}

type ServerOptions struct {
	Options
	Address string
}

func NewServerOptions(address string) ServerOptions {
	base := Options{}
	base.SetCompressor(compressor.Raw)
	base.SetSerializer(serializer.Json)
	return ServerOptions{Options: base, Address: address}
}

func NewClientOptions(address string) ClientOptions {
	base := Options{}
	base.SetCompressor(compressor.Raw)
	base.SetSerializer(serializer.Json)
	return ClientOptions{Options: base, Address: address}
}
