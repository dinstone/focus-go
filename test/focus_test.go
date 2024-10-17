package test

import (
	"errors"
	"log"
	"net"
	"net/rpc"
	"reflect"
	"testing"

	"github.com/dinstone/focus-go/focus"
	"github.com/dinstone/focus-go/focus/options"
	"github.com/dinstone/focus-go/focus/serializer"

	"github.com/dinstone/focus-go/focus/compressor"
	"github.com/dinstone/focus-go/test/json"
	"github.com/dinstone/focus-go/test/protobuf"
	"github.com/stretchr/testify/assert"
)

func init() {
}

// test client synchronously call
func client_call(t *testing.T, c compressor.Compressor) {
	server := createServer(c)
	defer server.Close()

	// client
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := focus.NewClient(conn, options.WithCompressor(c))
	defer client.Close()

	// case
	type expect struct {
		reply *protobuf.ArithResponse
		err   error
	}
	cases := []struct {
		client         *focus.Client
		name           string
		serviceMenthod string
		arg            *protobuf.ArithRequest
		expect         expect
	}{
		{
			client:         client,
			name:           "test-1",
			serviceMenthod: "ArithService.Add",
			arg:            &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 25},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-2",
			serviceMenthod: "ArithService.Sub",
			arg:            &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 15},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-3",
			serviceMenthod: "ArithService.Mul",
			arg:            &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 100},
				err:   nil,
			},
		},
		{
			client:         client,
			name:           "test-4",
			serviceMenthod: "ArithService.Div",
			arg:            &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 4},
			},
		},
		{
			client,
			"test-5",
			"ArithService.Div",
			&protobuf.ArithRequest{A: 20, B: 0},
			expect{
				&protobuf.ArithResponse{},
				rpc.ServerError("divided is zero"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &protobuf.ArithResponse{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}

func createServer(c compressor.Compressor) net.Listener {
	// server
	lis, err := net.Listen("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	server := focus.NewServer(options.WithCompressor(c))
	err = server.Register(new(protobuf.ArithService))
	if err != nil {
		log.Fatal(err)
	}
	go server.Serve(lis)

	return lis
}

// TestClient_Call test client synchronously call
func TestClient_Call(t *testing.T) {
	client_call(t, compressor.Raw)
}

// TestClient_AsyncCall test client asynchronously call
func TestClient_AsyncCall(t *testing.T) {
	server := createServer(nil)
	defer server.Close()

	// client
	conn, err := net.Dial("tcp", ":8008")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := focus.NewClient(conn)
	defer client.Close()

	type expect struct {
		reply *protobuf.ArithResponse
		err   error
	}
	cases := []struct {
		client        *focus.Client
		name          string
		serviceMethod string
		arg           *protobuf.ArithRequest
		expect        expect
	}{
		{
			client:        client,
			name:          "test-1",
			serviceMethod: "ArithService.Add",
			arg:           &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 25},
			},
		},
		{
			client:        client,
			name:          "test-2",
			serviceMethod: "ArithService.Sub",
			arg:           &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 15},
			},
		},
		{
			client:        client,
			name:          "test-3",
			serviceMethod: "ArithService.Mul",
			arg:           &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 100},
			},
		},
		{
			client:        client,
			name:          "test-4",
			serviceMethod: "ArithService.Div",
			arg:           &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 4},
			},
		},
		{
			client,
			"test-5",
			"ArithService.Div",
			&protobuf.ArithRequest{A: 20, B: 0},
			expect{
				&protobuf.ArithResponse{},
				rpc.ServerError("divided is zero"),
			},
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			reply := &protobuf.ArithResponse{}
			call := cs.client.AsyncCall(cs.serviceMethod, cs.arg, reply)
			err := <-call
			assert.Equal(t, true, reflect.DeepEqual(cs.expect.reply.C, reply.C))
			assert.Equal(t, cs.expect.err, err.Error)
		})
	}
}

// TestNewClientWithSnappyCompress test snappy comressor
func TestNewClientWithSnappyCompress(t *testing.T) {
	client_call(t, compressor.Snappy)
}

// TestNewClientWithGzipCompress test gzip comressor
func TestNewClientWithGzipCompress(t *testing.T) {
	client_call(t, compressor.Gzip)
}

// TestNewClientWithZlibCompress test zlib compressor
func TestNewClientWithZlibCompress(t *testing.T) {
	client_call(t, compressor.Zlib)
}

// TestServer_Register .
func TestServer_Register(t *testing.T) {
	server := focus.NewServer()
	err := server.RegisterName("ArithService", new(protobuf.ArithService))
	assert.Equal(t, nil, err)
	err = server.Register(new(protobuf.ArithService))
	assert.Equal(t, errors.New("rpc: service already defined: ArithService"), err)
}

// TestNewClientWithSerializer .
func TestNewClientWithSerializer(t *testing.T) {
	// server
	lis, err := net.Listen("tcp", ":8010")
	if err != nil {
		log.Fatal(err)
	}

	server := focus.NewServer(options.WithSerializer(serializer.Json))
	err = server.Register(new(json.TestService))
	if err != nil {
		log.Fatal(err)
	}
	go server.Serve(lis)

	// client
	conn, err := net.Dial("tcp", ":8010")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := focus.NewClient(conn, options.WithSerializer(serializer.Json))
	defer client.Close()

	type expect struct {
		reply *json.Response
		err   error
	}
	cases := []struct {
		client         *focus.Client
		name           string
		serviceMenthod string
		arg            *json.Request
		expect         expect
	}{
		{
			client:         client,
			name:           "test-1",
			serviceMenthod: "TestService.Add",
			arg:            &json.Request{A: 20, B: 5},
			expect: expect{
				reply: &json.Response{C: 25},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &json.Response{}
			err := c.client.Call(c.serviceMenthod, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}
