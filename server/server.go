package server

import (
	"fmt"
	"net"

	"github.com/EnemigoPython/go-getit/runtime"
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
			fmt.Printf("Message bytes: % x\n", buf[:n])
			c.Write([]byte("Hello back!"))
		}(conn)
	}
}
