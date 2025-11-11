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

	requestBytes := runtime.Frame(request)
	if runtime.Config.Debug {
		log.Printf("Request bytes: % x\n", requestBytes)
	}

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
		rawBytes := buf[:n]
		if runtime.Config.Debug {
			log.Printf("Raw bytes: % x\n", rawBytes)
		}
		done := false
		for frame := range runtime.FrameChannel(rawBytes) {
			response := runtime.DecodeResponse(frame)
			if runtime.Config.Debug {
				fmt.Println(response)
			}

			// don't read stream done to stdout
			if response.GetStatus() != runtime.StreamDone {
				// read all other responses
				fmt.Println(response.DataPayload())
			}

			if !request.IsStream() || response.GetStatus() != runtime.Ok {
				done = true
				break
			}
		}
		if done {
			break
		}
	}
}
