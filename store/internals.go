package store

import (
	"encoding/binary"
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
	keyLen := int(b[1])
	key := string(b[2 : 2+keyLen])
	fmt.Println(key)
	dataType := int(b[34])
	if dataType == 1 {
		valLen := int(b[35])
		val := string(b[36 : 36+valLen])
		fmt.Println(val)
	} else {
		val := int32(binary.BigEndian.Uint32(b[35:39]))
		fmt.Println(val)
	}
	x := 0
	return x, nil
}
