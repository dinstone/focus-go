package focus

import (
	"log"
	"net"
	"net/rpc"

	"github.com/dinstone/focus-go/focus/compressor"
	"github.com/dinstone/focus-go/focus/options"

	"github.com/dinstone/focus-go/focus/codec"
	"github.com/dinstone/focus-go/focus/serializer"
)

// Server rpc server based on net/rpc implementation
type Server struct {
	*rpc.Server
	options options.Options
}

func NewServer(opts ...options.Option) *Server {
	options := options.Options{
		Serializer: serializer.Proto,
		Compressor: compressor.Raw,
	}
	for _, option := range opts {
		option(&options)
	}

	server := &Server{&rpc.Server{}, options}
	return server
}

// Register register rpc function
func (s *Server) Register(rcvr interface{}) error {
	return s.Server.Register(rcvr)
}

// RegisterName register the rpc function with the specified name
func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return s.Server.RegisterName(name, rcvr)
}

// Serve start service
func (s *Server) Serve(lis net.Listener) {
	log.Printf("focus server started on: %s", lis.Addr().String())
	for {
		conn, err := lis.Accept()
		if err != nil {
			continue
		}
		go s.Server.ServeCodec(codec.NewServerCodec(conn, s.options.Serializer, s.options.Compressor))
	}
}
