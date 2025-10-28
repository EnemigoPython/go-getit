package client

import (
	"fmt"
	"net"

	"github.com/EnemigoPython/go-getit/runtime"
)

func MakeRequest() {
	conn, err := net.Dial("tcp", runtime.SocketAddress())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Send a message
	message := "Hello server!"
	_, err = conn.Write([]byte(message))
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
