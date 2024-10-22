package test

import (
	"errors"
	"log"
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
	// server
	server := createServer(c)
	defer server.Close()

	// client
	x := options.NewClientOptions(":8008")
	x.SetCompressor(c)
	x.SetSerializer(serializer.Protobuf)
	client := focus.NewClient(x)
	defer client.Close()

	// case
	type expect struct {
		reply *protobuf.ArithResponse
		err   error
	}
	cases := []struct {
		client  *focus.Client
		name    string
		service string
		method  string
		arg     *protobuf.ArithRequest
		expect  expect
	}{
		{
			client:  client,
			name:    "test-1",
			service: "ArithService",
			method:  "Add",
			arg:     &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 25},
				err:   nil,
			},
		},
		{
			client:  client,
			name:    "test-2",
			service: "ArithService",
			method:  "Sub",
			arg:     &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 15},
				err:   nil,
			},
		},
		{
			client:  client,
			name:    "test-3",
			service: "ArithService",
			method:  "Mul",
			arg:     &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 100},
				err:   nil,
			},
		},
		{
			client:  client,
			name:    "test-4",
			service: "ArithService",
			method:  "Div",
			arg:     &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 4},
			},
		},
		{
			client,
			"test-5",
			"ArithService",
			"Div",
			&protobuf.ArithRequest{A: 20, B: 0},
			expect{
				&protobuf.ArithResponse{},
				errors.New("(301)divided is zero"),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &protobuf.ArithResponse{}
			err := c.client.Call(c.service, c.method, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}

func createServer(c compressor.Compressor) *focus.Server {
	x := options.NewServerOptions(":8008")
	if c != nil {
		x.SetCompressor(c)
	}
	x.SetSerializer(serializer.Protobuf)
	server := focus.NewServer(x)
	err := server.Register(new(protobuf.ArithService))
	if err != nil {
		log.Fatal(err)
	}
	server.Start()

	return server
}

// TestClient_AsyncCall test client asynchronously call
func TestClient_AsyncCall(t *testing.T) {
	server := createServer(nil)
	defer server.Close()

	x := options.NewClientOptions(":8008")
	x.SetSerializer(serializer.Protobuf)
	client := focus.NewClient(x)
	defer client.Close()

	type expect struct {
		reply *protobuf.ArithResponse
		err   error
	}
	cases := []struct {
		client  *focus.Client
		name    string
		service string
		method  string
		arg     *protobuf.ArithRequest
		expect  expect
	}{
		{
			client:  client,
			name:    "test-1",
			service: "ArithService",
			method:  "Add",
			arg:     &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 25},
			},
		},
		{
			client:  client,
			name:    "test-2",
			service: "ArithService",
			method:  "Sub",
			arg:     &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 15},
			},
		},
		{
			client:  client,
			name:    "test-3",
			service: "ArithService",
			method:  "Mul",
			arg:     &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 100},
			},
		},
		{
			client:  client,
			name:    "test-4",
			service: "ArithService",
			method:  "Div",
			arg:     &protobuf.ArithRequest{A: 20, B: 5},
			expect: expect{
				reply: &protobuf.ArithResponse{C: 4},
			},
		},
		{
			client,
			"test-5",
			"ArithService",
			"Div",
			&protobuf.ArithRequest{A: 20, B: 0},
			expect{
				&protobuf.ArithResponse{},
				errors.New("(301)divided is zero"),
			},
		},
	}
	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			reply := &protobuf.ArithResponse{}
			call := cs.client.AsyncCall(cs.service, cs.method, cs.arg, reply)
			err := <-call
			assert.Equal(t, true, reflect.DeepEqual(cs.expect.reply.C, reply.C))
			assert.Equal(t, cs.expect.err, err.Error)
		})
	}
}

// TestClient_Call test client synchronously call
func TestClient_Call(t *testing.T) {
	client_call(t, compressor.Raw)
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
	server := focus.NewServer(options.ServerOptions{})
	err := server.RegisterName("ArithService", new(protobuf.ArithService))
	assert.Equal(t, nil, err)
	err = server.Register(new(protobuf.ArithService))
	assert.Equal(t, errors.New("rpc: service already defined: ArithService"), err)
}

// TestNewClientWithSerializer .
func TestNewClientWithSerializer(t *testing.T) {
	// server
	server := focus.NewServer(options.NewServerOptions(":8010"))
	err := server.Register(new(json.TestService))
	if err != nil {
		log.Fatal(err)
	}
	server.Start()

	// var wg sync.WaitGroup
	// wg.Add(1)
	// go func() {
	// 	// do something
	// 	wg.Done()
	// }()
	// wg.Wait()

	// client
	client := focus.NewClient(options.NewClientOptions(":8010"))
	defer client.Close()

	type expect struct {
		reply *json.Response
		err   error
	}
	cases := []struct {
		client  *focus.Client
		name    string
		service string
		method  string
		arg     *json.Request
		expect  expect
	}{
		{
			client:  client,
			name:    "test-1",
			service: "TestService",
			method:  "Add",
			arg:     &json.Request{A: 20, B: 5},
			expect: expect{
				reply: &json.Response{C: 25},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply := &json.Response{}
			err := c.client.Call(c.service, c.method, c.arg, reply)
			assert.Equal(t, true, reflect.DeepEqual(c.expect.reply.C, reply.C))
			assert.Equal(t, c.expect.err, err)
		})
	}
}
