package store

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/EnemigoPython/go-getit/runtime"
)

const entrySize int64 = 66       // number of bytes in file entry encoding
const maxEntrySpace int64 = 4200 // hash & file size limit
const seed int64 = 0xFACE        // random seed

type _storeMetadata struct {
	size       int64
	entrySpace int64
	entries    int64
}

var storeMetadata _storeMetadata

func readEntryBytes(file *os.File) int64 {
	// read first 4 bytes to get number of entries
	buf := make([]byte, 4)
	_, err := file.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	entries := int32(binary.BigEndian.Uint32(buf))
	return int64(entries)
}

func updateEntryBytes(file *os.File, update int64) {
	storeMetadata.entries += update
	file.Seek(0, io.SeekStart)
	binary.Write(file, binary.BigEndian, int32(storeMetadata.entries))
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

func readEntry(index int64, file *os.File) (decodedEntry, error) {
	file.Seek(index, io.SeekStart)
	buf := make([]byte, entrySize)
	n, err := file.Read(buf)
	if runtime.Config.Debug {
		fmt.Printf("Entry bytes: % x\n", buf)
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
	file.Seek(-index, io.SeekCurrent)
	return decoded, nil
}
