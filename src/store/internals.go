package store

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/EnemigoPython/go-getit/src/runtime"
)

const entrySize int64 = 66       // number of bytes in file entry encoding
const maxEntrySpace int64 = 5000 // hash & file size limit
const seed int64 = 0xFACE        // random seed
const maxCollisions = 3          // maximum permitted collisions
const streamBufferSize = 100     // size of stream channel
const workerCount = 10           // number of workers for stream

var keysFilter = []runtime.Status{runtime.NotFound}

type _storeMetadata struct {
	size       int64
	entrySpace int64
	entries    int64
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
		// new store; write empty metadata
		newMetaBytes := make([]byte, entrySize)
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

func hashKey(key string) (res int64) {
	for i, r := range key {
		res += seed
		res += int64((i + 1) * int(r))
		res <<= 1
		res ^= seed
		res <<= 2
		res %= maxEntrySpace
	}
	res += 1
	return
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
	fp.Seek(index, io.SeekStart)
	buf := make([]byte, entrySize)
	n, err := fp.Read(buf)
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
