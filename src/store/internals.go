package store

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/EnemigoPython/go-getit/src/runtime"
	"github.com/EnemigoPython/go-getit/src/types"
)

const entrySize int64 = 66     // number of bytes in file entry encoding
const minEntrySpace int64 = 50 // default hash & file size limit
const maxCollisions = 3        // maximum permitted collisions
const streamBufferSize = 100   // size of stream channel
const workerCount = 10         // number of workers for stream

var notFoundFilter = []runtime.Status{runtime.NotFound}

type _storeMetadata struct {
	size       int64 // size in bytes
	entrySpace int64 // current entry space
	entries    int64 // number of entries
}

var storeMetadata _storeMetadata

var mutex sync.RWMutex

func getReadPointer() (*os.File, error) {
	filePath := runtime.Config.StorePath
	fp, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	mutex.RLock()
	return fp, nil
}

func getReadWritePointer() (*os.File, error) {
	filePath := runtime.Config.StorePath
	fp, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	mutex.Lock()
	return fp, nil
}

func freeLock()  { mutex.Unlock() }
func freeRLock() { mutex.RUnlock() }

func readMetaBytes(fp *os.File) int64 {
	// read first 4 bytes to get number of entries
	buf := make([]byte, 4)
	_, err := fp.Read(buf)
	if err != nil {
		// new store; write empty metadata + min entry space
		minSize := (minEntrySpace * entrySize) + entrySize
		newMetaBytes := make([]byte, minSize)
		fp.Write(newMetaBytes)
	}
	entries := int32(binary.BigEndian.Uint32(buf))
	return int64(entries)
}

func updateEntryBytes(fp *os.File, update int64) {
	storeMetadata.entries += update
	fp.Seek(0, io.SeekStart)
	binary.Write(fp, binary.BigEndian, int32(storeMetadata.entries))
}

func entryIndex(i int64) int64 {
	return i * entrySize
}

// Implements DJB2 hashing
func hashKey(key string) int64 {
	var hash uint64 = 5381
	for _, r := range key {
		hash = ((hash << 5) + hash) + uint64(r)
	}
	return (int64(hash) % storeMetadata.entrySpace) + 1
}

type DecodeFileError struct {
	errorStr string
}

func (e DecodeFileError) Error() string {
	return fmt.Sprintf("Error decoding file; %s", e.errorStr)
}

type valueType int

const (
	typeInt valueType = iota
	typeString
)

type decodedEntry struct {
	IsSet     bool
	Key       string
	ValueType valueType
	Int       int
	Str       string
	Index     int64
}

func decodeFileBytes(b []byte) (decodedEntry, error) {
	if b[0] == 0 {
		return decodedEntry{IsSet: false}, nil
	}
	keyLen := int(b[1])
	key := string(b[2 : 2+keyLen])
	dataType := int(b[33])
	if dataType == 1 {
		valLen := int(b[34])
		val := string(b[35 : 35+valLen])
		return decodedEntry{
			IsSet:     true,
			Key:       key,
			ValueType: typeString,
			Str:       val,
		}, nil
	} else {
		val := int32(binary.BigEndian.Uint32(b[34:38]))
		return decodedEntry{
			IsSet:     true,
			Key:       key,
			ValueType: typeInt,
			Int:       int(val),
		}, nil
	}
}

func readEntry(index int64, fp *os.File) (decodedEntry, error) {
	buf := make([]byte, entrySize)
	n, err := fp.ReadAt(buf, index)
	if runtime.Config.Debug {
		log.Printf("Entry bytes: % x\n", buf)
	}
	if err != nil {
		return decodedEntry{}, err
	}
	if n < int(entrySize) {
		return decodedEntry{}, DecodeFileError{errorStr: "Insufficient bytes"}
	}
	decoded, err := decodeFileBytes(buf)
	if err != nil {
		return decodedEntry{}, err
	}
	decoded.Index = index
	return decoded, nil
}

func resolveEntry(index int64, fp *os.File, key string) (decodedEntry, error) {
	for range maxCollisions {
		decoded, err := readEntry(index, fp)
		if err != nil {
			log.Printf("Error; hit end of file at index %d\n", index)
			return decodedEntry{}, DecodeFileError{errorStr: "EOF"}
		}
		if !decoded.IsSet || decoded.Key == key {
			return decoded, nil
		}
		if runtime.Config.Debug {
			log.Printf(
				"Collision between keys %s and %s at index %d\n",
				key,
				decoded.Key,
				index,
			)
		}
		index += entrySize
	}
	log.Printf("Error; maximum search depth exceeded at %d for %s\n", index, key)
	return decodedEntry{}, DecodeFileError{errorStr: "Maximum search depth"}
}

// Overwrite the data record of an entry without modifying other bits
//
// Assumes the key has already been checked against the file index
func overwriteData[T types.IntOrString](index int64, fp *os.File, data T) {
	// seek to data section of index
	fp.Seek(index+int64(33), io.SeekStart)
	buf := new(bytes.Buffer)
	switch d := any(data).(type) {
	case int:
		runtime.WriteIntBytes(buf, d, true)
	case string:
		runtime.WriteStringBytes(buf, d, true)
	}
	fp.Write(buf.Bytes())
}
