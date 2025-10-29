package client

import (
	"fmt"
	"net"

	"github.com/EnemigoPython/go-getit/runtime"
)

func MakeRequest(message runtime.Message) {
	conn, err := net.Dial("tcp", runtime.SocketAddress())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	bytes := message.GetMessageBytes()

	// Send a message
	_, err = conn.Write(bytes)
	if err != nil {
		panic(err)
	}
	fmt.Println("Sent:", message)

	// Read the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println("Received:", string(buf[:n]))
}
