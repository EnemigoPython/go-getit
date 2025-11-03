package client

import (
	"fmt"
	"log"
	"net"

	"github.com/EnemigoPython/go-getit/runtime"
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

	// Read the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	response := runtime.DecodeResponse(buf[:n])
	if runtime.Config.Debug {
		fmt.Println(response)
	}
	fmt.Println(response.DataPayload())
}
