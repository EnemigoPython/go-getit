package client

import (
	"fmt"
	"log"
	"net"

	"github.com/EnemigoPython/go-getit/src/runtime"
)

func MakeRequest(request runtime.Request) {
	conn, err := net.Dial("tcp", runtime.SocketAddress())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	requestBytes := request.EncodeRequest()

	// Send a message
	_, err = conn.Write(requestBytes)
	if err != nil {
		log.Fatal(err)
	}
	if runtime.Config.Debug {
		fmt.Println(request)
	}

	for {
		// Read the response
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatal(err)
		}
		responseBytes := buf[:n]
		if runtime.Config.Debug {
			log.Printf("Response bytes: % x\n", responseBytes)
		}
		response := runtime.DecodeResponse(responseBytes)
		if runtime.Config.Debug {
			fmt.Println(response)
		}

		// don't read stream done to stdout
		if response.GetStatus() != runtime.StreamDone {
			// read all other responses
			fmt.Println(response.DataPayload())
		}

		if !request.IsStream() || response.GetStatus() != runtime.Ok {
			break
		}
	}
}
