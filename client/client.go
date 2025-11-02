package client

import (
	"fmt"
	"net"

	"github.com/EnemigoPython/go-getit/runtime"
)

func MakeRequest(request runtime.Request) {
	conn, err := net.Dial("tcp", runtime.SocketAddress())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	bytes := request.EncodeRequest()

	// Send a message
	_, err = conn.Write(bytes)
	if err != nil {
		panic(err)
	}
	if runtime.Config.Debug {
		fmt.Println(request)
	}

	// Read the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println("Response:", string(buf[:n]))
}
