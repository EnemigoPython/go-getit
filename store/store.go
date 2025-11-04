package store

import (
	"fmt"
	"io"
	"os"

	"github.com/EnemigoPython/go-getit/runtime"
)

func OpenStore() error {
	filename := runtime.FileName()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	info, _ := os.Stat(filename)
	fileSize := info.Size()
	entries := readEntryBytes(file)
	storeMetadata = _storeMetadata{
		size:       int64(fileSize),
		entrySpace: (int64(fileSize) / entrySize) - 1,
		entries:    entries,
	}
	fmt.Printf("Using store '%s'\n", filename)
	return nil
}

func store(request runtime.Request) runtime.Response {
	fp, err := getReadWritePointer()
	if err != nil {
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	defer fp.Close()
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		fmt.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		fp.Seek(0, io.SeekEnd)
		paddingLen := index - storeMetadata.size
		paddedBytes := make([]byte, paddingLen)
		fp.Write(paddedBytes)
		storeMetadata.size += paddingLen
		storeMetadata.entrySpace += paddingLen / entrySize
		updateEntryBytes(fp, 1)
	} else {
		decoded, err := readEntry(index, fp)
		if err != nil {
			return runtime.ConstructResponse(request, runtime.ServerError, 0)
		}
		if !decoded.IsSet {
			updateEntryBytes(fp, 1)
		}
	}
	fp.Seek(index, io.SeekStart)
	fp.Write(request.EncodeFileBytes())
	storeMetadata.size += entrySize
	storeMetadata.entrySpace++
	freeLock()
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func load(request runtime.Request) runtime.Response {
	fp, err := getReadPointer()
	if err != nil {
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	defer fp.Close()
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		fmt.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		return runtime.ConstructResponse(request, runtime.NotFound, 0)
	}
	decoded, err := readEntry(index, fp)
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

func clear(request runtime.Request) runtime.Response {
	fp, err := getReadWritePointer()
	if err != nil {
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	defer fp.Close()
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		fmt.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		return runtime.ConstructResponse(request, runtime.Ok, 0)
	}
	decoded, err := readEntry(index, fp)
	if err != nil {
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	fp.Seek(index, io.SeekStart)
	fp.Write([]byte{0}) // unset header byte
	if decoded.IsSet {
		// if the entry was previously set decrement the entries counter
		updateEntryBytes(fp, -1)
	}
	freeLock()
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func clearAll(request runtime.Request) runtime.Response {
	fp, err := getReadWritePointer()
	if err != nil {
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	defer fp.Close()
	fp.Truncate(entrySize)
	storeMetadata.size = entrySize
	storeMetadata.entrySpace = 0
	updateEntryBytes(fp, -storeMetadata.entries)
	freeLock()
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func count(request runtime.Request) runtime.Response {
	return runtime.ConstructResponse(
		request,
		runtime.Ok,
		int(storeMetadata.entries),
	)
}

func ProcessRequest(request runtime.Request) runtime.Response {
	switch request.GetAction() {
	case runtime.Store:
		return store(request)
	case runtime.Load:
		return load(request)
	case runtime.Clear:
		return clear(request)
	case runtime.ClearAll:
		return clearAll(request)
	case runtime.Count:
		return count(request)
	}
	panic("Unreachable")
}
