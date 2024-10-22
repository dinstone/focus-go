package client

import (
	"reflect"
	"testing"

	"github.com/dinstone/focus-go/focus"
	"github.com/dinstone/focus-go/focus/options"
	"github.com/dinstone/focus-go/focus/serializer"
	pb "github.com/dinstone/focus-go/test/protobuf"
	"github.com/stretchr/testify/assert"
)

func TestFocusProtobuf(t *testing.T) {
	x := options.NewClientOptions(":8010")
	x.SetSerializer(serializer.Protobuf)
	client := focus.NewClient(x)
	defer client.Close()

	expect := &pb.ArithResponse{C: 25}
	service := "ArithService"
	menthod := "Add"
	arg := &pb.ArithRequest{A: 20, B: 5}
	reply := &pb.ArithResponse{}
	err := client.Call(service, menthod, arg, reply)

	assert.Equal(t, nil, err)
	assert.Equal(t, true, reflect.DeepEqual(expect.C, reply.C))
}

func TestFocusJson(t *testing.T) {
	x := options.NewClientOptions(":3344")
	client := focus.NewClient(x)
	defer client.Close()

	expect := "hi dinstone, from go"
	service := "com.dinstone.focus.example.DemoService"
	menthod := "hello"
	arg := "dinstone, from go"
	var reply string
	err := client.Call(service, menthod, arg, &reply)

	assert.Equal(t, nil, err)
	assert.Equal(t, true, reflect.DeepEqual(expect, reply))
}
