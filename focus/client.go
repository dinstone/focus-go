package focus

import (
	"io"
	"net/rpc"

	"github.com/dinstone/focus-go/focus/compressor"
	"github.com/dinstone/focus-go/focus/options"
	"github.com/dinstone/focus-go/focus/serializer"

	"github.com/dinstone/focus-go/focus/codec"
)

// Client rpc client based on net/rpc implementation
type Client struct {
	*rpc.Client
}

// NewClient Create a new rpc client
func NewClient(conn io.ReadWriteCloser, opts ...options.Option) *Client {
	options := options.Options{
		Serializer: serializer.Proto,
		Compressor: compressor.Raw,
	}
	for _, option := range opts {
		option(&options)
	}
	return &Client{rpc.NewClientWithCodec(
		codec.NewClientCodec(conn, options.Compressor, options.Serializer))}
}

// Call synchronously calls the rpc function
func (c *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return c.Client.Call(serviceMethod, args, reply)
}

// AsyncCall asynchronously calls the rpc function and returns a channel of *rpc.Call
func (c *Client) AsyncCall(serviceMethod string, args interface{}, reply interface{}) chan *rpc.Call {
	return c.Go(serviceMethod, args, reply, nil).Done
}
