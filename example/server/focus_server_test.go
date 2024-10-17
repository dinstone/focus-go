package server

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/dinstone/focus-go/focus"
	"github.com/dinstone/focus-go/focus/options"
	"github.com/dinstone/focus-go/focus/serializer"
	js "github.com/dinstone/focus-go/test/json"
	pb "github.com/dinstone/focus-go/test/protobuf"
)

func TestJsonServer(t *testing.T) {
	lis, err := net.Listen("tcp", ":9010")
	if err != nil {
		log.Fatal(err)
	}
	server := focus.NewServer(options.WithSerializer(serializer.Json))
	err = server.Register(new(js.TestService))
	if err != nil {
		log.Fatal(err)
	}
	go server.Serve(lis)

	fmt.Println("service is on 9010")
	time.Sleep(time.Duration(2) * time.Hour)
}

func TestProtoServer(t *testing.T) {
	lis, err := net.Listen("tcp", ":8010")
	if err != nil {
		log.Fatal(err)
	}
	server := focus.NewServer(options.WithSerializer(serializer.Proto))
	err = server.Register(new(pb.ArithService))
	if err != nil {
		log.Fatal(err)
	}
	go server.Serve(lis)

	fmt.Println("service is on 8010")
	time.Sleep(time.Duration(2) * time.Hour)
}
