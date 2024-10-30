package focus

import (
	"errors"
	"go/token"
	"log"
	"net"
	"reflect"
	"sync"

	"github.com/dinstone/focus-go/focus/compressor"
	"github.com/dinstone/focus-go/focus/options"
	"github.com/dinstone/focus-go/focus/protocol"
	"github.com/dinstone/focus-go/focus/serializer"
	"github.com/dinstone/focus-go/focus/transport"
)

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type StatusType struct {
	Code    int
	Message string
}

type methodType struct {
	sync.Mutex // protects counters
	method     reflect.Method
	ArgType    reflect.Type
	ReplyType  reflect.Type
	numCalls   uint
}

type serviceType struct {
	name   string                 // name of service
	rcvr   reflect.Value          // receiver of methods for the service
	typ    reflect.Type           // type of the receiver
	method map[string]*methodType // registered methods
}

type Server struct {
	options    options.ServerOptions
	listener   net.Listener
	serviceMap sync.Map
}

func NewServer(opts options.ServerOptions) *Server {
	return &Server{options: opts, serviceMap: sync.Map{}}
}

// Register register rpc function
func (s *Server) Register(rcvr interface{}) error {
	return s.register(rcvr, "", false)
}

// RegisterName register the rpc function with the specified name
func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return s.register(rcvr, name, true)
}

func (server *Server) register(rcvr any, name string, useName bool) error {
	s := new(serviceType)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := name
	if !useName {
		sname = reflect.Indirect(s.rcvr).Type().Name()
	}
	if sname == "" {
		s := "rpc.Register: no service name for type " + s.typ.String()
		log.Print(s)
		return errors.New(s)
	}
	if !useName && !token.IsExported(sname) {
		s := "rpc.Register: type " + sname + " is not exported"
		log.Print(s)
		return errors.New(s)
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(s.typ, false)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PointerTo(s.typ), false)
		if len(method) != 0 {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type"
		}
		log.Print(str)
		return errors.New(str)
	}

	if _, dup := server.serviceMap.LoadOrStore(sname, s); dup {
		return errors.New("rpc: service already defined: " + sname)
	}
	return nil
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}

// suitableMethods returns suitable Rpc methods of typ. It will log
// errors if logErr is true.
func suitableMethods(typ reflect.Type, logErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if !method.IsExported() {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			if logErr {
				log.Printf("rpc.Register: method %q has %d input parameters; needs exactly three\n", mname, mtype.NumIn())
			}
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			if logErr {
				log.Printf("rpc.Register: argument type of method %q is not exported: %q\n", mname, argType)
			}
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Pointer {
			if logErr {
				log.Printf("rpc.Register: reply type of method %q is not a pointer: %q\n", mname, replyType)
			}
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if logErr {
				log.Printf("rpc.Register: reply type of method %q is not exported: %q\n", mname, replyType)
			}
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			if logErr {
				log.Printf("rpc.Register: method %q has %d output parameters; needs exactly one\n", mname, mtype.NumOut())
			}
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			if logErr {
				log.Printf("rpc.Register: return type of method %q is %q, must be error\n", mname, returnType)
			}
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", s.options.Address)
	if err != nil {
		log.Fatal(err)
	}
	s.listener = listener

	log.Printf("focus server started on: %s", listener.Addr().String())

	go func() {
		s.accept()
	}()
}

func (s *Server) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Print("focus server accept:", err.Error())
			log.Fatal(err)
		}
		go s.process(transport.NewConnection(conn))
	}
}

func (s *Server) process(conn *transport.Connection) {
	for {
		// read request from connetion
		request, err := conn.ReadMessage()
		if err != nil {
			log.Println("rpc read error:", err)
			break
		}

		// heartbeat message
		if request.MsgType == 0 {
			request.Status = 1
			conn.WriteMessage(request)

			continue
		}

		// not request message: response or notice
		if request.MsgType != 1 {
			continue
		}

		// Look up the svciv and call method.
		// Decode the argument value.
		// if true, need to indirect before calling.
		// argv guaranteed to be a pointer now.
		// Invoke the method, providing a new value for the reply.
		// The return value for the method is an error.
		compressor := s.options.GetCompressor()
		serializer := s.options.GetSerializer()

		replyv, status := s.invoke(request, compressor, serializer)

		respone := new(protocol.Message)
		respone.Version = request.Version
		respone.Sequence = request.Sequence
		respone.MsgType = 2

		respone.Headers = make(protocol.Headers, 2)

		if status.Code != 0 {
			respone.Status = int16(status.Code)
			respone.Content = []byte(status.Message)
		} else {
			respone.Status = 0
			respone.Headers["serializer.type"] = serializer.Type()
			respBody, err := serializer.Encode(replyv.Interface())
			if err != nil {
				respone.Status = 201
				respone.Content = []byte("encode reply error:" + err.Error())
			} else {
				respone.Content, _ = compressor.Encode(respBody)
			}
		}

		conn.WriteMessage(respone)
	}

	conn.Close()
}

func (s *Server) invoke(request *protocol.Message, compressor compressor.Compressor, serializer serializer.Serializer) (*reflect.Value, StatusType) {
	var status StatusType

	serviceName := request.Headers["call.service"]
	methodName := request.Headers["call.method"]
	svciv, ok := s.serviceMap.Load(serviceName)
	if !ok {
		status = StatusType{202, "rpc: can't find service " + serviceName}
		return nil, status
	}
	stype := svciv.(*serviceType)
	mtype := stype.method[methodName]
	if mtype == nil {
		status = StatusType{203, "rpc: can't find method " + methodName}
		return nil, status
	}

	var argref reflect.Value
	argIsValue := false
	if mtype.ArgType.Kind() == reflect.Pointer {
		argref = reflect.New(mtype.ArgType.Elem())
	} else {
		argref = reflect.New(mtype.ArgType)
		argIsValue = true
	}

	// decode arg
	content, err := compressor.Decode(request.Content)
	if err != nil {
		status = StatusType{201, err.Error()}
	}
	err = serializer.Decode(content, argref.Interface())
	if err != nil {
		status = StatusType{201, err.Error()}
	}
	if argIsValue {
		argref = argref.Elem()
	}

	replyv := reflect.New(mtype.ReplyType.Elem())
	switch mtype.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(mtype.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(mtype.ReplyType.Elem(), 0, 0))
	}

	function := mtype.method.Func
	// invoke
	returnValues := function.Call([]reflect.Value{stype.rcvr, argref, replyv})
	// return value process
	errInter := returnValues[0].Interface()
	if errInter != nil {
		status = StatusType{301, errInter.(error).Error()}
		return nil, status
	}
	return &replyv, status
}

func (s *Server) Close() {
	s.listener.Close()
}
