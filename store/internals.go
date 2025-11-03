package store

import (
	"encoding/binary"
	"fmt"
)

const entrySize int64 = 66       // number of bytes in file encoding
const maxEntrySpace int64 = 4200 // hash & file size limit
const seed int64 = 0xFACE        // random seed

type _storeMetadata struct {
	size       int64
	entrySpace int64
}

var storeMetadata _storeMetadata

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

type decodedValue struct {
	Type valueType
	Int  int
	Str  string
}

func decodeFileBytes(b []byte) (decodedValue, error) {
	if b[0] == 0 {
		return decodedValue{}, DecodeFileError{errorStr: "Unset entry"}
	}
	keyLen := int(b[1])
	key := string(b[2 : 2+keyLen])
	if key == "" {
		// TODO: colision detection
	}
	dataType := int(b[33])
	if dataType == 1 {
		valLen := int(b[34])
		val := string(b[35 : 35+valLen])
		return decodedValue{Type: typeString, Str: val}, nil
	} else {
		val := int32(binary.BigEndian.Uint32(b[34:38]))
		return decodedValue{Type: typeInt, Int: int(val)}, nil
	}
}
