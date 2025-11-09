package runtime

import (
	"bytes"
	"encoding/binary"
)

// Write encoded bytes for an entry key with optional padding
func WriteKeyBytes(buf *bytes.Buffer, key string, pad bool) {
	keyLen := len(key)
	buf.WriteByte(byte(keyLen)) // number of bytes
	buf.Write([]byte(key))
	if pad {
		paddedBytes := make([]byte, maxStringLen-keyLen)
		buf.Write(paddedBytes)
	}
}

// Write encoded int to buffer with optional padding
func WriteIntBytes(buf *bytes.Buffer, i int, pad bool) {
	buf.WriteByte(byte(0)) // type of data: int
	binary.Write(buf, binary.BigEndian, int32(i))
	if pad {
		paddedBytes := make([]byte, 28)
		buf.Write(paddedBytes)
	}
}

// Write encoded string to buffer with optional padding
func WriteStringBytes(buf *bytes.Buffer, s string, pad bool) {
	dataLen := len(s)
	buf.WriteByte(byte(1))      // type of data: string
	buf.WriteByte(byte(len(s))) // number of bytes
	buf.Write([]byte(s))
	if pad {
		paddedBytes := make([]byte, maxStringLen-dataLen)
		buf.Write(paddedBytes)
	}
}
