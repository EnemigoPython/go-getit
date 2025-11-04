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
	entries := readEntryBytes(file)
	storeMetadata = _storeMetadata{
		size:       int64(fileSize),
		entrySpace: (int64(fileSize) / entrySize) - 1,
		entries:    entries,
	}
	fmt.Printf("Opened store '%s'\n", filename)
	return file, nil
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
		updateEntryBytes(file, 1)
	} else {
		decoded, err := readEntry(index, file)
		if err != nil {
			return runtime.ConstructResponse(request, runtime.ServerError, 0)
		}
		if !decoded.IsSet {
			updateEntryBytes(file, 1)
		}
	}
	file.Seek(index, io.SeekStart)
	file.Write(request.EncodeFileBytes())
	storeMetadata.size += entrySize
	storeMetadata.entrySpace++
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func load(request runtime.Request, file *os.File) runtime.Response {
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		fmt.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		return runtime.ConstructResponse(request, runtime.NotFound, 0)
	}
	decoded, err := readEntry(index, file)
	if err != nil {
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	if !decoded.IsSet {
		return runtime.ConstructResponse(request, runtime.NotFound, 0)
	}
	switch decoded.ValueType {
	case typeInt:
		return runtime.ConstructResponse(request, runtime.Ok, decoded.Int)
	case typeString:
		return runtime.ConstructResponse(request, runtime.Ok, decoded.Str)
	}
	panic("Unreachable")
}

func clear(request runtime.Request, file *os.File) runtime.Response {
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		fmt.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		return runtime.ConstructResponse(request, runtime.Ok, 0)
	}
	decoded, err := readEntry(index, file)
	if err != nil {
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	file.Seek(index, io.SeekStart)
	file.Write([]byte{0}) // unset header byte
	if decoded.IsSet {
		// if the entry was previously set decrement the entries counter
		updateEntryBytes(file, -1)
	}
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func clearAll(request runtime.Request, file *os.File) runtime.Response {
	file.Truncate(entrySize)
	storeMetadata.size = entrySize
	storeMetadata.entrySpace = 0
	updateEntryBytes(file, -storeMetadata.entries)
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func count(request runtime.Request) runtime.Response {
	return runtime.ConstructResponse(request, runtime.Ok, int(storeMetadata.entries))
}

func ProcessRequest(request runtime.Request, file *os.File) runtime.Response {
	switch request.GetAction() {
	case runtime.Store:
		return store(request, file)
	case runtime.Load:
		return load(request, file)
	case runtime.Clear:
		return clear(request, file)
	case runtime.ClearAll:
		return clearAll(request, file)
	case runtime.Count:
		return count(request)
	}
	panic("Unreachable")
}
