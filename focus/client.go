package focus

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/dinstone/focus-go/focus/options"
	"github.com/dinstone/focus-go/focus/protocol"
	"github.com/dinstone/focus-go/focus/transport"
)

// Client rpc client based on net/tcp implementation
type Client struct {
	options options.ClientOptions
	connect *transport.Connection

	reqMutex sync.Mutex // protects following
	// request  Request

	mutex    sync.Mutex // protects following
	seq      int32
	pending  map[int32]*Call
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

// NewClient Create a new rpc client
func NewClient(opts options.ClientOptions) *Client {
	conn, err := net.Dial("tcp", opts.Address)
	if err != nil {
		log.Fatal(err)
	}

	client := &Client{options: opts, pending: make(map[int32]*Call), connect: transport.NewConnection(conn)}
	go client.input()
	return client
}

func (client *Client) input() {
	var err error
	var response *protocol.Message
	for err == nil {
		response, err = client.connect.ReadMessage()
		if err != nil {
			break
		}
		// heartbeat message
		if response.MsgType == 0 {
			continue
		}
		// call's response message
		seq := response.Sequence
		client.mutex.Lock()
		call := client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()

		if call != nil {
			if response.Status == 0 {
				response.Content, _ = client.options.GetCompressor().Decode(response.Content)
				client.options.GetSerializer().Decode(response.Content, call.Reply)
			} else {
				errmsg := string(response.Content)
				call.Error = fmt.Errorf("(%d)%s", response.Status, errmsg)
			}
			call.done()
		}
	}

	// Terminate pending calls.
	client.reqMutex.Lock()
	client.mutex.Lock()
	client.shutdown = true
	closing := client.closing
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
	client.mutex.Unlock()
	client.reqMutex.Unlock()
	if err != io.EOF && !closing {
		log.Println("rpc: client protocol error:", err)
	}
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here. It is the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
		log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

// Call synchronously calls the rpc function
func (c *Client) Call(service string, method string, args interface{}, reply interface{}) error {
	call := <-c.Go(service, method, args, reply, make(chan *Call, 1)).Done
	return call.Error
}

// AsyncCall asynchronously calls the rpc function and returns a channel
func (c *Client) AsyncCall(service string, method string, args interface{}, reply interface{}) chan *Call {
	return c.Go(service, method, args, reply, nil).Done
}

func (client *Client) Go(service string, method string, args any, reply any, done chan *Call) *Call {
	call := new(Call)
	call.Service = service
	call.Method = method
	call.Args = args
	call.Reply = reply
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel. If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	client.send(call)
	return call
}

var ErrShutdown = errors.New("connection is shut down")

func (client *Client) send(call *Call) {
	client.reqMutex.Lock()
	defer client.reqMutex.Unlock()

	// Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		client.mutex.Unlock()
		call.Error = ErrShutdown
		call.done()
		return
	}
	seq := client.seq
	client.seq++
	client.pending[seq] = call
	client.mutex.Unlock()

	// Encode and send the request.
	request := new(protocol.Message)
	request.Version = 1
	request.MsgType = 1
	request.Sequence = seq

	serializer := client.options.GetSerializer()
	content, err := serializer.Encode(call.Args)
	if err != nil {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
		return
	}
	compressor := client.options.GetCompressor()
	request.Content, err = compressor.Encode(content)
	if err != nil {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
		return
	}

	request.Headers = make(protocol.Headers, 2)
	request.Headers["serializer.type"] = serializer.Type()
	request.Headers["call.service"] = call.Service
	request.Headers["call.method"] = call.Method

	err = client.connect.WriteMessage(request)
	if err != nil {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (client *Client) Close() {
	client.mutex.Lock()
	if client.closing {
		client.mutex.Unlock()
	}
	client.closing = true
	client.mutex.Unlock()
	client.connect.Close()
}

// Call represents an active RPC.
type Call struct {
	Service string     // The name of the service to call.
	Method  string     // The name of the method to call.
	Args    any        // The argument to the function (*struct).
	Reply   any        // The reply from the function (*struct).
	Error   error      // After completion, the error status.
	Done    chan *Call // Receives *Call when Go is complete.
}
