package server

import (
	"fmt"
	"log"
	"net"

	"github.com/EnemigoPython/go-getit/runtime"
	"github.com/EnemigoPython/go-getit/store"
)

func Run() {
	if runtime.Config.Debug {
		fmt.Println("Running in debug mode")
	}
	ln, _ := net.Listen("tcp", runtime.SocketAddress())
	fmt.Println("Listening on port", runtime.Config.Port)
	defer ln.Close()

	file, err := store.OpenStore()
	if err != nil {
		log.Fatal(err)
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
			response := store.ProcessRequest(request, file)
			fmt.Println(response)
			responseBytes := response.EncodeResponse()
			if runtime.Config.Debug {
				fmt.Printf("Response bytes: % x\n", responseBytes)
			}
			c.Write(responseBytes)
		}(conn)
	}
}
