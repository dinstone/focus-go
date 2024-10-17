# Focus-go

**Focus** is the next generation cross language lightweight RPC framework. It can quickly and easily develop microservice applications, which greatly simplifies RPC programming.

**Focus-go** is the go language implementation of the Focus.

## Install

- install `protoc` at first : http://github.com/google/protobuf/releases
- install `protoc-gen-go` and `protoc-gen-focus` :

```shell
go install github.com/golang/protobuf/protoc-gen-go
go install github.com/dinstone/focus-go/protoc-gen-focus
```

## Quick Start

1. create a demo project and import the focus package:

```shell
> go mod init demo
> go get github.com/dinstone/focus-go
```

2. under the path of the project, create a protobuf file `arith.proto`:

```protobuf
syntax = "proto3";

package protobuf;
option go_package="/protobuf";

// ArithService Defining Computational Digital Services
service ArithService {
  // Add addition
  rpc Add(ArithRequest) returns (ArithResponse);
  // Sub subtraction
  rpc Sub(ArithRequest) returns (ArithResponse);
  // Mul multiplication
  rpc Mul(ArithRequest) returns (ArithResponse);
  // Div division
  rpc Div(ArithRequest) returns (ArithResponse);
}

message ArithRequest {
  int32 a = 1;
  int32 b = 2;
}

message ArithResponse {
  int32 c = 1;
}
```

3. using `protoc` to generate code:

```shell
> protoc --focus_out=. arith.proto --go_out=. arith.proto
```

at this time, two files will be generated in the directory `protobuf`: `arith.pb.go` and `arith.svr.go`

4. implement the ArithService in the file `arith.svr.go` :

```go
package protobuf

import "errors"

// ArithService Defining Computational Digital Services
type ArithService struct{}

// Add addition
func (this *ArithService) Add(args *ArithRequest, reply *ArithResponse) error {
	reply.C = args.A + args.B
	return nil
}

// Sub subtraction
func (this *ArithService) Sub(args *ArithRequest, reply *ArithResponse) error {
	reply.C = args.A - args.B
	return nil
}

// Mul multiplication
func (this *ArithService) Mul(args *ArithRequest, reply *ArithResponse) error {
	reply.C = args.A * args.B
	return nil
}

// Div division
func (this *ArithService) Div(args *ArithRequest, reply *ArithResponse) error {
	if args.B == 0 {
		return errors.New("divided is zero")
	}
	reply.C = args.A / args.B
	return nil
}
```

## Server

under the path of the project, we create a file named `focus_server.go`, create a focus server and publish service.

```go
package main

import (
	"demo/protobuf"
	"log"
	"net"

	"github.com/dinstone/focus-go"
)

func main() {
	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatal(err)
	}

	server := focus.NewServer()
	server.RegisterName("ArithService", new(protobuf.ArithService))
	server.Serve(lis)
}
```

## Client

under the path of the project, we create a file named `focus_client.go`, create a focus client and call it synchronously with the `Add` function:

```go
import (
	"demo/protobuf"
	"github.com/dinstone/focus-go"
...

conn, err := net.Dial("tcp", ":8082")
if err != nil {
	log.Fatal(err)
}
defer conn.Close()
client := focus.NewClient(conn)
resq := protobuf.ArithRequest{A: 20, B: 5}
resp := protobuf.ArithResponse{}
err = client.Call("ArithService.Add", &resq, &resp)
log.Printf("Arith.Add(%v, %v): %v ,Error: %v", resq.A, resq.B, resp.C, err)
```
you can also call asynchronously, which will return a channel of type *rpc.Call:
```go

result := client.AsyncCall("ArithService.Add", &resq, &resp)
select {
case call := <-result:
	log.Printf("Arith.Add(%v, %v): %v ,Error: %v", resq.A, resq.B, resp.C, call.Error)
case <-time.After(100 * time.Microsecond):
	log.Fatal("time out")
}
```
of course, you can also compress with three supported formats `gzip`, `snappy`, `zlib`:
```go
import (
    "github.com/dinstone/focus-go"
    "github.com/dinstone/focus-go/options"
    "github.com/dinstone/focus-go/serializer"
	"github.com/dinstone/focus-go/compressor"
)

...
client := focus.NewClient(conn, options.WithCompress(compressor.Gzip))

```
## Custom Serializer
If you want to customize the serializer, you must implement the `Serializer` interface:
```go
type Serializer interface {
    Marshal(message interface{}) ([]byte, error)
    Unmarshal(data []byte, message interface{}) error
    Type() string
}
```
`JsonSerializer` is a serializer based Json:
```go
type JsonSerializer struct{}

func (_ JsonSerializer) Marshal(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}

func (_ JsonSerializer) Unmarshal(data []byte, message interface{}) error {
	return json.Unmarshal(data, message)
}
```
now, we can create a HelloService with the following code:
```go
type HelloRequest struct {
	Req string `json:"req"`
}

type HelloResponce struct {
	Resp string `json:"resp"`
}

type HelloService struct{}

func (_ *HelloService) SayHello(args *HelloRequest, reply *HelloResponce) error {
	reply.Resp = args.Req
	return nil
}

```
finally, we need to set the serializer on the focus server:
```go
func main() {
	lis, err := net.Listen("tcp", ":8082")
	if err != nil {
		log.Fatal(err)
	}

	server := focus.NewServer(options.WithSerializer(serializer.Json))
	server.Register(new(HelloService))
	server.Serve(lis)
}
```

Remember that when the focus client calls the service, it also needs to set the serializer:
```go
focus.NewClient(conn, options.WithSerializer(serializer.Json))
```
