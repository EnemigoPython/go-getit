package server

import (
	"fmt"
	"net"

	"github.com/EnemigoPython/go-getit/runtime"
	"github.com/EnemigoPython/go-getit/store"
)

func Run() {
	ln, _ := net.Listen("tcp", runtime.SocketAddress())
	fmt.Println("Listening on port", runtime.Config.Port)
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 1024)
			n, _ := c.Read(buf)
			requestBytes := buf[:n]
			// TODO: show only if in debug mode
			fmt.Printf("Request bytes: % x\n", requestBytes)
			request := runtime.DecodeRequest(requestBytes)
			fmt.Println(request)
			store.ProcessRequest(request)
			c.Write([]byte("Hello back!"))
		}(conn)
	}
}
