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

	file, err := store.OpenStore()
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for {
		conn, _ := ln.Accept()
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 1024)
			n, _ := c.Read(buf)
			requestBytes := buf[:n]
			if runtime.Config.Debug {
				fmt.Printf("Request bytes: % x\n", requestBytes)
			}
			request := runtime.DecodeRequest(requestBytes)
			fmt.Println(request)
			store.ProcessRequest(request, file)
			c.Write([]byte("Hello back!"))
		}(conn)
	}
}
