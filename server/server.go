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
			messageBytes := buf[:n]
			// TODO: show only if in debug mode
			fmt.Printf("Message bytes: % x\n", messageBytes)
			message := runtime.DecodeMessage(messageBytes)
			fmt.Println(message)
			c.Write([]byte("Hello back!"))
		}(conn)
	}
}
