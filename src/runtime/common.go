package runtime

import (
	"bytes"
	"encoding/binary"
)

// Handles message boundary detection for requests/responses
type Framer interface {
	Encode() []byte
}

func Frame(f Framer) []byte {
	b := f.Encode()
	msg := make([]byte, 2+len(b))
	binary.BigEndian.PutUint16(msg, uint16(len(b)))
	copy(msg[2:], b)
	return msg
}

func FrameChannel(b []byte) <-chan []byte {
	out := make(chan []byte)
	bytesLen := uint16(len(b))
	var index uint16
	go func() {
		defer close(out)
		for index < bytesLen {
			header := binary.BigEndian.Uint16(b[index : index+1])
			out <- b[index+2 : index+header+2]
			index += header + 2
		}
	}()
	return out
}

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
