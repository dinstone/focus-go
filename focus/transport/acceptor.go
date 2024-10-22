package transport

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Acceptor struct {
	listener   net.Listener
	serviceMap *sync.Map
}

func (a *Acceptor) Bind(port int, services *sync.Map) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	a.listener = listener
	a.serviceMap = services
	log.Printf("focus server started on: %s", listener.Addr().String())

	for {
		conn, err := a.listener.Accept()
		if err != nil {
			continue
		}
		go a.process(NewConnection(conn))
	}
}

func (a *Acceptor) process(conn *Connection) {

	conn.Close()
}
