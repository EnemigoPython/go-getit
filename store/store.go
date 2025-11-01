package store

import (
	"fmt"
	"os"

	"github.com/EnemigoPython/go-getit/runtime"
)

func OpenStore() {
	filename := runtime.Config.StoreName
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	fmt.Println("File opened or created successfully:", filename)
}

func store(request runtime.Request) runtime.Response {
	r := runtime.ConstructResponse(runtime.Status(0))
	return r
}

func load(request runtime.Request) runtime.Response {
	r := runtime.ConstructResponse(runtime.Status(0))
	return r
}

func clear(request runtime.Request) runtime.Response {
	r := runtime.ConstructResponse(runtime.Status(0))
	return r
}

func ProcessRequest(request runtime.Request) {
	var response runtime.Response
	switch request.GetAction() {
	case runtime.Store:
		response = store(request)
	case runtime.Load:
		response = load(request)
	case runtime.Clear:
		response = clear(request)
	}
	fmt.Println(response)
}
