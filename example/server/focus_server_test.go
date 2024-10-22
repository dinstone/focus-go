package server

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/dinstone/focus-go/focus"
	"github.com/dinstone/focus-go/focus/options"
	"github.com/dinstone/focus-go/focus/serializer"
	js "github.com/dinstone/focus-go/test/json"
	pb "github.com/dinstone/focus-go/test/protobuf"
)

func TestJsonServer(t *testing.T) {
	server := focus.NewServer(options.NewServerOptions(":9010"))
	err := server.Register(new(js.TestService))
	if err != nil {
		log.Fatal(err)
	}
	go server.Start()
	time.Sleep(time.Duration(2) * time.Hour)
}

func TestProtoServer(t *testing.T) {
	x := options.NewServerOptions(":8010")
	x.SetSerializer(serializer.Protobuf)
	server := focus.NewServer(x)
	err := server.Register(new(pb.ArithService))
	if err != nil {
		log.Fatal(err)
	}
	go server.Start()

	fmt.Println("service is on 8010")
	time.Sleep(time.Duration(2) * time.Hour)
}
