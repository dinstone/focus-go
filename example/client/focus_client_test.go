package client

import (
	"github.com/dinstone/focus-go"
	"github.com/dinstone/focus-go/options"
	"github.com/dinstone/focus-go/serializer"
	pb "github.com/dinstone/focus-go/test/protobuf"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"reflect"
	"testing"
)

func TestFocusProtobuf(t *testing.T) {
	conn, err := net.Dial("tcp", ":3333")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := focus.NewClient(conn)
	defer client.Close()

	expect := &pb.ArithResponse{C: 25}
	serviceMenthod := "ArithService.Add"
	arg := &pb.ArithRequest{A: 20, B: 5}
	reply := &pb.ArithResponse{}
	err = client.Call(serviceMenthod, arg, reply)

	assert.Equal(t, nil, err)
	assert.Equal(t, true, reflect.DeepEqual(expect.C, reply.C))
}

func TestFocusJson(t *testing.T) {
	conn, err := net.Dial("tcp", ":3333")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := focus.NewClient(conn, options.WithSerializer(serializer.Json))
	defer client.Close()

	expect := "hi dinstone, from go"
	serviceMenthod := "com.dinstone.focus.example.DemoService.hello"
	arg := "dinstone, from go"
	var reply string
	err = client.Call(serviceMenthod, arg, &reply)

	assert.Equal(t, nil, err)
	assert.Equal(t, true, reflect.DeepEqual(expect, reply))
}
