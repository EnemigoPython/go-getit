package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/EnemigoPython/go-getit/src/runtime"
	"github.com/EnemigoPython/go-getit/src/store"
)

func handleConnection(ln net.Listener, c net.Conn) {
	defer c.Close()
	buf := make([]byte, 1024)
	n, _ := c.Read(buf)
	requestBytes := buf[:n]
	if runtime.Config.Debug {
		log.Printf("Request bytes: % x\n", requestBytes)
	}
	request := runtime.DecodeRequest(requestBytes)
	log.Println(request)

	// stream requests need to handle multiple responses
	if request.IsStream() {
		var endStream runtime.Response
		for response := range store.ProcessStreamRequest(request) {
			// on error or stream end, write other responses first
			if response.GetStatus() != runtime.Ok {
				endStream = response
				continue
			}
			log.Println(response)
			responseBytes := response.Encode()
			// responseBytes := runtime.EncodeDelimited(response)
			if runtime.Config.Debug {
				log.Printf("Response bytes: % x\n", responseBytes)
			}

			// write to socket
			c.Write(responseBytes)
		}

		// now log & send captured end stream
		log.Println(endStream)
		responseBytes := endStream.Encode()
		if runtime.Config.Debug {
			log.Printf("Response bytes: % x\n", responseBytes)
		}
		c.Write(responseBytes)
		return
	}

	// non-streamed response
	response := store.ProcessRequest(request)
	log.Println(response)
	responseBytes := response.Encode()
	if runtime.Config.Debug {
		log.Printf("Response bytes: % x\n", responseBytes)
	}

	// write to socket
	c.Write(responseBytes)

	// exit if command was to shut down
	if request.GetAction() == runtime.Exit {
		ln.Close()
		return
	}
}

func Run() {
	if runtime.Config.Debug {
		fmt.Println("Running in debug mode")
	}
	configureLogger()
	defer shutdown()
	ln, err := net.Listen("tcp", runtime.SocketAddress())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on port", runtime.Config.Port)
	defer ln.Close()

	// notify the program when an OS shutdown signal occurs
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig // block until the shutdown signal and then exit gracefully
		ln.Close()
	}()

	err = store.OpenStore()
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// explicit exit; otherwise panic
			return
		}
		go handleConnection(ln, conn)
	}
}
