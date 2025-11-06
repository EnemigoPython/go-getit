package store

import (
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"sync"

	"github.com/EnemigoPython/go-getit/src/runtime"
)

func OpenStore() error {
	filePath := runtime.Config.StorePath
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	info, _ := os.Stat(filePath)
	fileSize := info.Size()
	entries := readMetaBytes(file)
	storeMetadata = _storeMetadata{
		size:       int64(fileSize),
		entrySpace: (int64(fileSize) / entrySize) - 1,
		entries:    entries,
	}
	log.Printf("Using store '%s'\n", filePath)
	return nil
}

func store(request runtime.Request, fp *os.File) runtime.Response {
	hash := hashKey(request.GetKey())
	index := entryIndex(hash)
	var code int // 0=overwrite value, 1=new value
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
		code = 1
	} else {
		decoded, err := resolveEntry(index, fp, request.GetKey())
		if err != nil {
			return runtime.ConstructResponse(request, runtime.ServerError, 0)
		}
		if !decoded.IsSet {
			updateEntryBytes(fp, 1)
			code = 1
		}
		index = decoded.Index
	}
	fp.Seek(index, io.SeekStart)
	fp.Write(request.EncodeFileBytes())
	storeMetadata.size += entrySize
	storeMetadata.entrySpace++
	return runtime.ConstructResponse(request, runtime.Ok, code)
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

func keys(request runtime.Request, fp *os.File, i int) runtime.Response {
	index := entryIndex(int64(i + 1))
	if storeMetadata.size < index {
		return runtime.ConstructResponse(request, runtime.StreamDone, 0)
	}
	decodedEntry, err := readEntry(index, fp)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return runtime.ConstructResponse(request, runtime.ServerError, 0)
	}
	if decodedEntry.IsSet {
		return runtime.ConstructResponse(request, runtime.Ok, decodedEntry.Key)
	}
	return runtime.ConstructResponse(request, runtime.NotFound, 0)
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

func streamReadOperation(
	f func(runtime.Request, *os.File, int) runtime.Response,
	request runtime.Request,
	statusFilter []runtime.Status,
	out chan<- runtime.Response,
) {
	stop := make(chan struct{})
	var once sync.Once
	nextIndex := make(chan int)

	go func() {
		fp, err := getReadPointer()
		if err != nil {
			out <- runtime.ConstructResponse(request, runtime.ServerError, 0)
			return
		}
		defer fp.Close()
		defer freeRLock()

		// feed next index to channel in loop
		go func() {
			for i := 0; ; i++ {
				select {
				case nextIndex <- i:
				case <-stop:
					close(nextIndex)
					return
				}
			}
		}()

		var wg sync.WaitGroup

		for range workerCount {
			wg.Go(func() {
				for idx := range nextIndex {
					response := f(request, fp, idx)
					if !slices.Contains(statusFilter, response.GetStatus()) {
						out <- response
					}

					// shut down if signal is stream done
					if response.GetStatus() == runtime.StreamDone {
						once.Do(func() { close(stop) })
						return
					}
				}
			})
		}

		// exit after all workers done
		wg.Wait()
		close(out)
	}()
}

func ProcessStreamRequest(request runtime.Request) <-chan runtime.Response {
	out := make(chan runtime.Response, streamBufferSize)

	switch request.GetAction() {
	case runtime.Keys:
		go streamReadOperation(keys, request, keysFilter, out)
	default:
		panic("Unreachable")
	}

	return out
}
