package store

import (
	"io"
	"log"
	"os"

	"github.com/EnemigoPython/go-getit/src/runtime"
)

func OpenStore() error {
	filename := runtime.FileName()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	info, _ := os.Stat(filename)
	fileSize := info.Size()
	entries := readMetaBytes(file)
	storeMetadata = _storeMetadata{
		size:       int64(fileSize),
		entrySpace: (int64(fileSize) / entrySize) - 1,
		entries:    entries,
	}
	log.Printf("Using store '%s'\n", filename)
	return nil
}

func store(request runtime.Request, fp *os.File) runtime.Response {
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		log.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		fp.Seek(0, io.SeekEnd)
		extraPadding := entrySize * 3 // for collision cases
		paddingLen := (index - storeMetadata.size) + extraPadding
		paddedBytes := make([]byte, paddingLen)
		fp.Write(paddedBytes)
		storeMetadata.size += paddingLen
		storeMetadata.entrySpace += paddingLen / entrySize
		updateEntryBytes(fp, 1)
	} else {
		decoded, err := resolveEntry(index, fp, request.GetKey())
		if err != nil {
			return runtime.ConstructResponse(request, runtime.ServerError, 0)
		}
		if !decoded.IsSet {
			updateEntryBytes(fp, 1)
		}
		index = decoded.Index
	}
	fp.Seek(index, io.SeekStart)
	fp.Write(request.EncodeFileBytes())
	storeMetadata.size += entrySize
	storeMetadata.entrySpace++
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func load(request runtime.Request, fp *os.File) runtime.Response {
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		log.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		return runtime.ConstructResponse(request, runtime.NotFound, 0)
	}
	decoded, err := resolveEntry(index, fp, request.GetKey())
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

func clear(request runtime.Request, fp *os.File) runtime.Response {
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	if runtime.Config.Debug {
		log.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		return runtime.ConstructResponse(request, runtime.Ok, 0)
	}
	decoded, err := resolveEntry(index, fp, request.GetKey())
	if err != nil {
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	index = decoded.Index
	fp.Seek(index, io.SeekStart)
	fp.Write([]byte{0}) // unset header byte
	if decoded.IsSet {
		// if the entry was previously set decrement the entries counter
		updateEntryBytes(fp, -1)
	}
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func clearAll(request runtime.Request, fp *os.File) runtime.Response {
	fp.Truncate(entrySize)
	storeMetadata.size = entrySize
	storeMetadata.entrySpace = 0
	updateEntryBytes(fp, -storeMetadata.entries)
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func count(request runtime.Request) runtime.Response {
	return runtime.ConstructResponse(
		request,
		runtime.Ok,
		int(storeMetadata.entries),
	)
}

func exit(request runtime.Request) runtime.Response {
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func readOperation(
	f func(runtime.Request, *os.File) runtime.Response,
	request runtime.Request,
) runtime.Response {
	fp, err := getReadPointer()
	if err != nil {
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	defer fp.Close()
	defer freeRLock()
	return f(request, fp)
}

func writeOperation(
	f func(runtime.Request, *os.File) runtime.Response,
	request runtime.Request,
) runtime.Response {
	fp, err := getReadWritePointer()
	if err != nil {
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	defer fp.Close()
	defer freeLock()
	return f(request, fp)
}

func ProcessRequest(request runtime.Request) runtime.Response {
	switch request.GetAction() {
	case runtime.Store:
		return writeOperation(store, request)
	case runtime.Load:
		return readOperation(load, request)
	case runtime.Clear:
		return writeOperation(clear, request)
	case runtime.ClearAll:
		return writeOperation(clearAll, request)
	case runtime.Count:
		return count(request)
	case runtime.Exit:
		return exit(request)
	}
	panic("Unreachable")
}
