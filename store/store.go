package store

import (
	"fmt"
	"os"

	"github.com/EnemigoPython/go-getit/runtime"
)

const entrySize uint64 = 64

type _storeMetadata struct {
	size    uint64
	entries uint64
}

var storeMetadata _storeMetadata

func entryIndex(i uint64) uint64 {
	return i * entrySize
}

func hashKey(key string) (res uint64) {
	for i, r := range key {
		res += uint64((i + 1) * int(r))
	}
	return
}

func OpenStore() (*os.File, error) {
	filename := runtime.FileName()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	info, _ := os.Stat(filename)
	fileSize := info.Size()
	storeMetadata = _storeMetadata{
		size:    uint64(fileSize),
		entries: uint64(fileSize) / entrySize,
	}
	fmt.Printf("Opened store '%s'\n", filename)
	return file, err
}

func store(request runtime.Request) runtime.Response {
	r := runtime.ConstructResponse(request, runtime.Ok, 0)
	return r
}

func load(request runtime.Request) runtime.Response {
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		fmt.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		r := runtime.ConstructResponse(request, runtime.NotFound, 0)
		return r
	}
	r := runtime.ConstructResponse(request, runtime.Ok, 0)
	return r
}

func clear(request runtime.Request) runtime.Response {
	r := runtime.ConstructResponse(request, runtime.Ok, "A")
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
