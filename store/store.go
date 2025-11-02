package store

import (
	"fmt"
	"io"
	"os"

	"github.com/EnemigoPython/go-getit/runtime"
)

func OpenStore() (*os.File, error) {
	filename := runtime.FileName()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	info, _ := os.Stat(filename)
	fileSize := info.Size()
	storeMetadata = _storeMetadata{
		size:       int64(fileSize),
		entrySpace: int64(fileSize) / entrySize,
	}
	fmt.Printf("Opened store '%s'\n", filename)
	return file, err
}

func store(request runtime.Request, file *os.File) runtime.Response {
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		fmt.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		file.Seek(0, io.SeekEnd)
		paddingLen := index - storeMetadata.size
		paddedBytes := make([]byte, paddingLen)
		file.Write(paddedBytes)
		storeMetadata.size += paddingLen
		storeMetadata.entrySpace += paddingLen / entrySize
	} else {
		file.Seek(index, io.SeekStart)
	}
	file.Write(request.EncodeFileBytes())
	storeMetadata.size += entrySize
	storeMetadata.entrySpace++
	r := runtime.ConstructResponse(request, runtime.Ok, 0)
	return r
}

func load(request runtime.Request, file *os.File) runtime.Response {
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		fmt.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		r := runtime.ConstructResponse(request, runtime.NotFound, 0)
		return r
	}
	file.Seek(index, io.SeekStart)
	buf := make([]byte, entrySize)
	n, err := file.Read(buf)
	if runtime.Config.Debug {
		fmt.Printf("Entry bytes: % x\n", buf)
	}
	if err != nil || n < int(entrySize) {
		r := runtime.ConstructResponse(request, runtime.ServerError, 0)
		return r
	}
	r := runtime.ConstructResponse(request, runtime.Ok, 0)
	return r
}

func clear(request runtime.Request, file *os.File) runtime.Response {
	r := runtime.ConstructResponse(request, runtime.Ok, "A")
	return r
}

func ProcessRequest(request runtime.Request, file *os.File) {
	var response runtime.Response
	switch request.GetAction() {
	case runtime.Store:
		response = store(request, file)
	case runtime.Load:
		response = load(request, file)
	case runtime.Clear:
		response = clear(request, file)
	}
	fmt.Println(response)
}
