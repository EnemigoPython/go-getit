package store

import (
	"fmt"
	"io"
	"log"
	"math"
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
	minSize := (minTableSpace * entrySize) + entrySize
	entries := readMetaBytes(file, minSize)
	info, _ := os.Stat(filePath)
	fileSize := int64(info.Size())
	tableSpace := (fileSize / entrySize) - 1
	setRatio := float64(entries) / float64(tableSpace)
	storeMetadata = _storeMetadata{
		size:       fileSize,
		tableSpace: tableSpace,
		entries:    entries,
		setRatio:   setRatio,
		minSize:    minSize,
	}
	log.Printf("Using store '%s': %+v\n", filePath, storeMetadata)
	return nil
}

func store(request runtime.Request, fp *os.File) runtime.Response {
	hash := hashKey(request.GetKey(), storeMetadata.tableSpace)
	index := entryIndex(hash)
	var code int // 0=overwrite value, 1=new value
	if runtime.Config.Debug {
		log.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			"Index outside of file",
		)
	}
	decoded, err := resolveEntry(index, fp, request.GetKey())
	if err != nil {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	if !decoded.IsSet {
		updateEntryBytes(fp, 1)
		code = 1
		go checkResizeUp()
	}
	index = decoded.Index
	fp.WriteAt(request.EncodeFileBytes(), index)
	return runtime.ConstructResponse(request, runtime.Ok, code)
}

func arithmeticOperation(
	request runtime.Request,
	fp *os.File,
	a runtime.ArithmeticType,
) runtime.Response {
	hash := hashKey(request.GetKey(), storeMetadata.tableSpace)
	index := entryIndex(hash)
	if runtime.Config.Debug {
		log.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			"Index outside of file",
		)
	}
	decoded, err := resolveEntry(index, fp, request.GetKey())
	if err != nil {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	if !decoded.IsSet {
		return runtime.ConstructResponse(request, runtime.NotFound, 0)
	}
	switch decoded.ValueType {
	case typeInt:
		calculatedVal, err := request.ArithmeticOperation(a, decoded.Int)
		if err != nil {
			return runtime.ConstructResponse(
				request,
				runtime.ServerError,
				err.Error(),
			)
		}
		if calculatedVal < math.MinInt32 || calculatedVal > math.MaxInt32 {
			return runtime.ConstructResponse(
				request,
				runtime.InvalidRequest,
				"Operation causes overflow or underflow",
			)
		}
		overwriteData(index, fp, calculatedVal)
		return runtime.ConstructResponse(request, runtime.Ok, calculatedVal)
	case typeString:
		var errorMessage string
		if a == runtime.A_Add {
			errorMessage = "Cannot add to string"
		} else {
			errorMessage = "Cannot subtract from string"
		}
		return runtime.ConstructResponse(
			request,
			runtime.InvalidRequest,
			errorMessage,
		)
	}
	panic("Unreachable")
}

func add(request runtime.Request, fp *os.File) runtime.Response {
	return arithmeticOperation(request, fp, runtime.A_Add)
}

func sub(request runtime.Request, fp *os.File) runtime.Response {
	return arithmeticOperation(request, fp, runtime.A_Sub)
}

func load(request runtime.Request, fp *os.File) runtime.Response {
	hash := hashKey(request.GetKey(), storeMetadata.tableSpace)
	index := entryIndex(hash)
	if runtime.Config.Debug {
		log.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			"Index outside of file",
		)
	}
	decoded, err := resolveEntry(index, fp, request.GetKey())
	if err != nil {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
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
	hash := hashKey(request.GetKey(), storeMetadata.tableSpace)
	index := entryIndex(hash)
	if runtime.Config.Debug {
		log.Printf("Hash: %d, Index: %d\n", hash, index)
	}
	if storeMetadata.size < index {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			"Index outside of file",
		)
	}
	decoded, err := resolveEntry(index, fp, request.GetKey())
	if err != nil {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	index = decoded.Index
	fp.Seek(index, io.SeekStart)
	fp.Write([]byte{0}) // unset header byte
	if decoded.IsSet {
		// if the entry was previously set decrement the entries counter
		updateEntryBytes(fp, -1)
		go checkResizeDown()
		return runtime.ConstructResponse(request, runtime.Ok, 0)
	}
	return runtime.ConstructResponse(request, runtime.NotFound, 0)
}

func clearAll(request runtime.Request, fp *os.File) runtime.Response {
	fp.Truncate(storeMetadata.minSize)
	storeMetadata.size = storeMetadata.minSize
	storeMetadata.tableSpace = minTableSpace
	// format remaining table space
	formatLen := storeMetadata.minSize - entrySize
	buf := make([]byte, formatLen)
	fp.WriteAt(buf, entrySize)
	updateEntryBytes(fp, -storeMetadata.entries)
	return runtime.ConstructResponse(request, runtime.Ok, 0)
}

func keys(request runtime.Request, fp *os.File, i int) runtime.Response {
	index := entryIndex(int64(i + 1))
	if storeMetadata.size < index {
		return runtime.ConstructResponse(request, runtime.StreamDone, 0)
	}
	decoded, err := readEntry(index, fp, false)
	if err != nil && err != io.EOF {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	if decoded.IsSet {
		return runtime.ConstructResponse(request, runtime.Ok, decoded.Key)
	}
	return runtime.ConstructResponse(request, runtime.NotFound, 0)
}

func values(request runtime.Request, fp *os.File, i int) runtime.Response {
	index := entryIndex(int64(i + 1))
	if storeMetadata.size < index {
		return runtime.ConstructResponse(request, runtime.StreamDone, 0)
	}
	decoded, err := readEntry(index, fp, false)
	if err != nil && err != io.EOF {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	if decoded.IsSet {
		switch decoded.ValueType {
		case typeInt:
			return runtime.ConstructResponse(request, runtime.Ok, decoded.Int)
		case typeString:
			return runtime.ConstructResponse(request, runtime.Ok, decoded.Str)
		}
	}
	return runtime.ConstructResponse(request, runtime.NotFound, 0)
}

func items(request runtime.Request, fp *os.File, i int) runtime.Response {
	index := entryIndex(int64(i + 1))
	if storeMetadata.size < index {
		return runtime.ConstructResponse(request, runtime.StreamDone, 0)
	}
	decoded, err := readEntry(index, fp, false)
	if err != nil && err != io.EOF {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	if decoded.IsSet {
		var itemRow string
		switch decoded.ValueType {
		case typeInt:
			itemRow = fmt.Sprintf("%s %d", decoded.Key, decoded.Int)
		case typeString:
			itemRow = fmt.Sprintf("%s %s", decoded.Key, decoded.Str)
		}
		return runtime.ConstructResponse(request, runtime.Ok, itemRow)
	}
	return runtime.ConstructResponse(request, runtime.NotFound, 0)
}

func resize(request runtime.Request) runtime.Response {
	// we will free the read pointer manually
	fp, err := getReadPointer()
	if err != nil {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	newTableSpace, err := request.GetIntData()
	if err != nil {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	newSetRatio := float64(storeMetadata.entries) / float64(newTableSpace)
	// lenience on size down as it will only be applied when clearing keys
	if newSetRatio > sizeUpThreshold {
		return runtime.ConstructResponse(
			request,
			runtime.InvalidRequest,
			"Resize outside acceptable threshold",
		)
	}
	// create new file for overwrite
	filePath := runtime.Config.TempPath
	temp_fp, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			"Error opening temp file",
		)
	}
	defer temp_fp.Close()
	newFileSize := (int64(newTableSpace) * entrySize) + entrySize
	// format in case an artifact already existed
	err = temp_fp.Truncate(0)
	if err != nil {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	err = temp_fp.Truncate(newFileSize)
	if err != nil {
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	// write current entries to new file metadata
	updateEntryBytes(temp_fp, storeMetadata.entries)

	nextIndex := make(chan int64)
	resChannel := make(chan runtime.Response)
	var wg sync.WaitGroup
	var tempMutex sync.Mutex
	go func() {
		defer close(nextIndex)
		for i := entrySize; i < storeMetadata.size; i += entrySize {
			nextIndex <- i
		}
	}()
	for range workerCount {
		wg.Go(func() {
			for index := range nextIndex {
				decodedEntry, err := readEntry(index, fp, false)
				if err != nil {
					resChannel <- runtime.ConstructResponse(
						request,
						runtime.ServerError,
						err.Error(),
					)
				}
				if !decodedEntry.IsSet {
					continue
				}
				newHash := hashKey(decodedEntry.Key, int64(newTableSpace))
				newIndex := entryIndex(newHash)
				fmt.Println(decodedEntry.Key, newHash, newIndex)
				tempMutex.Lock()
				newDecodedEntry, err := resolveEntry(
					newIndex,
					temp_fp,
					decodedEntry.Key,
				)
				if err != nil {
					resChannel <- runtime.ConstructResponse(
						request,
						runtime.ServerError,
						err.Error(),
					)
				}
				newIndex = newDecodedEntry.Index
				temp_fp.WriteAt(request.EncodeFileBytes(), newIndex)
				tempMutex.Unlock()
			}
		})
	}
	go func() {
		wg.Wait()
		resChannel <- runtime.ConstructResponse(request, runtime.Ok, 0)
	}()
	response := <-resChannel
	// if no errors, replace with new file
	if response.GetStatus() == runtime.Ok {
		// close all file pointers & acquire write lock to rename
		info, erra := temp_fp.Stat()
		if erra != nil {
			panic(erra)
		}
		fmt.Println("Size:", info.Size(), "Expected:", newFileSize)
		temp_fp.Close()
		fp.Close()
		freeRLock()
		acquireLock()
		defer freeLock()
		err := os.Rename(runtime.Config.TempPath, runtime.Config.StoreName)
		if err != nil {
			return runtime.ConstructResponse(
				request,
				runtime.ServerError,
				err.Error(),
			)
		}
		testFp, _ := os.Open(runtime.Config.StorePath)
		info, _ = testFp.Stat()
		fmt.Println("AFTER RENAME SIZE:", info.Size())
		testFp.Close()
		storeMetadata.size = newFileSize
		storeMetadata.tableSpace = int64(newTableSpace)
		storeMetadata.setRatio = newSetRatio
	}
	return response
}

func count(request runtime.Request) runtime.Response {
	return runtime.ConstructResponse(
		request,
		runtime.Ok,
		int(storeMetadata.entries),
	)
}

func size(request runtime.Request) runtime.Response {
	return runtime.ConstructResponse(
		request,
		runtime.Ok,
		int(storeMetadata.size),
	)
}

func space(request runtime.Request) runtime.Response {
	switch request.GetKey() {
	case "current":
		return runtime.ConstructResponse(
			request,
			runtime.Ok,
			int(storeMetadata.tableSpace),
		)
	case "empty":
		emptyEntries := storeMetadata.tableSpace - storeMetadata.entries
		return runtime.ConstructResponse(
			request,
			runtime.Ok,
			int(emptyEntries),
		)
	}
	return runtime.ConstructResponse(request, runtime.ServerError, "Bad Verb")
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
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
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
		return runtime.ConstructResponse(
			request,
			runtime.ServerError,
			err.Error(),
		)
	}
	defer fp.Close()
	defer freeLock()
	return f(request, fp)
}

func ProcessRequest(request runtime.Request) runtime.Response {
	switch request.GetAction() {
	case runtime.Store:
		return writeOperation(store, request)
	case runtime.Add:
		return writeOperation(add, request)
	case runtime.Sub:
		return writeOperation(sub, request)
	case runtime.Load:
		return readOperation(load, request)
	case runtime.Clear:
		return writeOperation(clear, request)
	case runtime.ClearAll:
		return writeOperation(clearAll, request)
	case runtime.Resize:
		// uses a temp file so no need to block readers
		return resize(request)
	case runtime.Count:
		return count(request)
	case runtime.Size:
		return size(request)
	case runtime.Space:
		return space(request)
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
			out <- runtime.ConstructResponse(
				request,
				runtime.ServerError,
				err.Error(),
			)
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
		go streamReadOperation(keys, request, notFoundFilter, out)
	case runtime.Values:
		go streamReadOperation(values, request, notFoundFilter, out)
	case runtime.Items:
		go streamReadOperation(items, request, notFoundFilter, out)
	default:
		panic("Unreachable")
	}

	return out
}
