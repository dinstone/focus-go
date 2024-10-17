package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func StartServer() {
	listen, err := net.Listen("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println("net.listen error = ", err)
	}
	log.Printf("net.listen on %s", listen.Addr().String())
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("net.accept error = ", err)
			continue
		}
		go process(conn)
	}
}

func process(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		var buffer [1024]byte
		n, err := reader.Read(buffer[:])
		if err != nil {
			fmt.Println("conn.read error = ", err)
			break
		}

		receive := string(buffer[:n])
		fmt.Printf("server received data [%s]: %s \n", conn.RemoteAddr(), receive)

		_, err = conn.Write(buffer[:n])
		if err != nil {
			fmt.Println("conn.write error = ", err)
			break
		}
	}
	fmt.Println("conn.close : ", conn.RemoteAddr())
}

func main() {
	go StartServer()

	StartClient()
}

func StartClient() {
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println("net.Dial error : ", err)
		return
	}
	// 关闭连接
	defer conn.Close()
	// 键入数据
	inputReader := bufio.NewReader(os.Stdin)
	for {
		// 读取用户输入
		input, _ := inputReader.ReadString('\n')
		// 截断
		inputInfo := strings.Trim(input, "\r\n")
		// 读取到用户输入q 或者 Q 就退出
		if strings.ToUpper(inputInfo) == "Q" {
			return
		}
		// 将输入的数据发送给服务端
		_, err = conn.Write([]byte(inputInfo))
		if err != nil {
			return
		}
		buf := [1024]byte{}
		n, err := conn.Read(buf[:])
		if err != nil {
			fmt.Println("conn.Read error : ", err)
			return
		}
		fmt.Println(string(buf[:n]))
	}

}
