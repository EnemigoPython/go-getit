package store

import (
	"fmt"
)

const entrySize int64 = 68 // number of bytes in file encoding

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
		res += int64((i + 1) * (int(r) - 32))
	}
	return
}

type DecodeFileError struct {
	errorStr string
}

func (e DecodeFileError) Error() string {
	return fmt.Sprintf("Error decoding file; %s", e.errorStr)
}

func decodeFileBytes(b []byte) (int, error) {
	if b[0] == 0 {
		return 0, DecodeFileError{errorStr: "Empty entry"}
	}
	x := 0
	return x, nil
}
