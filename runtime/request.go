package runtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/EnemigoPython/go-getit/types"
)

type RequestParseError struct {
	errorStr string
}

func (e RequestParseError) Error() string {
	return fmt.Sprintf("Error parsing request; %s", e.errorStr)
}

type Action byte

const (
	Store Action = iota
	Load
	Clear
	ClearAll
)

func (a Action) String() string {
	return [...]string{"Store", "Load", "Clear", "ClearAll"}[a]
}

func (a Action) ToLower() string {
	return [...]string{"store", "load", "clear", "clearall"}[a]
}

func parseAction(s string) (Action, error) {
	switch strings.ToLower(s) {
	case "":
		return Action(0), RequestParseError{errorStr: "invalid action: <empty>"}
	case Store.ToLower():
		return Store, nil
	case Load.ToLower():
		return Load, nil
	case Clear.ToLower():
		return Clear, nil
	default:
		return Action(0), RequestParseError{errorStr: s}
	}
}

var requestCounter uint8

func generateId() uint8 {
	go func() {
		requestCounter++
	}()
	return requestCounter
}

const maxStringLen = 31

type request[T types.IntOrString] struct {
	action Action
	key    string
	data   T
	id     uint8
}

type Request interface {
	GetAction() Action
	GetKey() string
	GetId() uint8
	EncodeRequest() []byte
	EncodeFileBytes() []byte
}

func (r request[T]) GetAction() Action { return r.action }
func (r request[T]) GetKey() string    { return r.key }
func (r request[T]) GetId() uint8      { return r.id }

func (r request[T]) writeKeyBytes(buf *bytes.Buffer, pad bool) {
	keyLen := len(r.key)
	buf.WriteByte(byte(keyLen)) // number of bytes
	buf.Write([]byte(r.key))
	if pad {
		paddedBytes := make([]byte, maxStringLen-keyLen)
		buf.Write(paddedBytes)
	}
}

func (r request[T]) writeDataBytes(buf *bytes.Buffer, pad bool) {
	switch d := any(r.data).(type) {
	case int:
		buf.WriteByte(byte(0)) // type of data: int
		binary.Write(buf, binary.BigEndian, int32(d))
		if pad {
			paddedBytes := make([]byte, 28)
			buf.Write(paddedBytes)
		}
	case string:
		dataLen := len(d)
		buf.WriteByte(byte(1))      // type of data: string
		buf.WriteByte(byte(len(d))) // number of bytes
		buf.Write([]byte(d))
		if pad {
			paddedBytes := make([]byte, maxStringLen-dataLen)
			buf.Write(paddedBytes)
		}
	}
}

func (r request[T]) EncodeRequest() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(r.action))
	switch r.action {
	case Store:
		r.writeKeyBytes(buf, false)
		r.writeDataBytes(buf, false)
	case Load, Clear:
		r.writeKeyBytes(buf, false)
	case ClearAll:
		// no extra data fields needed
	}
	return buf.Bytes()
}

func (r request[T]) EncodeFileBytes() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(1)            // set first byte to signal stored
	r.writeKeyBytes(buf, true)  // write key with padding
	r.writeDataBytes(buf, true) // write data with padding
	return buf.Bytes()
}

func (r request[T]) String() string {
	var body string
	switch r.action {
	case Store:
		switch d := any(r.data).(type) {
		case int:
			body = fmt.Sprintf("%s[%s:%d]", r.action, r.key, d)
		case string:
			body = fmt.Sprintf("%s[%s:'%s']", r.action, r.key, d)
		default:
			panic("Unreachable")
		}
	case Load, Clear:
		body = fmt.Sprintf("%s[%s]", r.action, r.key)
	case ClearAll:
		body = r.action.String()
	default:
		panic("Unreachable")
	}
	return fmt.Sprintf("Request(%d)<%s>", r.id, body)
}

func ConstructRequest(args []string) (Request, error) {
	action, err := parseAction(args[0])
	if err != nil {
		return request[int]{}, err
	}
	var key string
	var data string
	switch action {
	case Store:
		if len(args) < 2 {
			return request[int]{}, RequestParseError{
				errorStr: "need 3 args for store",
			}
		}
		key = args[1]
		if len(key) > maxStringLen {
			return request[int]{}, RequestParseError{
				errorStr: fmt.Sprintf(
					"Key must be less than %d characters",
					maxStringLen,
				),
			}
		}
		data = args[2]
		if i, err := strconv.Atoi(data); err == nil {
			if i < math.MinInt32 || i > math.MaxInt32 {
				return request[int]{}, RequestParseError{
					errorStr: fmt.Sprintf(
						"invalid int data (must be %d-%d)",
						math.MinInt32,
						math.MaxInt32,
					),
				}
			}
			return request[int]{key: key, data: i, action: action}, nil
		}
		if len(data) > maxStringLen {
			return request[int]{}, RequestParseError{
				errorStr: fmt.Sprintf(
					"Data must be less than %d characters",
					maxStringLen,
				),
			}
		}
		return request[string]{key: key, data: data, action: action}, nil
	case Load:
		if len(args) < 1 {
			return request[int]{}, RequestParseError{
				errorStr: "need 2 args for load",
			}
		}
		key = args[1]
		if len(key) > maxStringLen {
			return request[int]{}, RequestParseError{
				errorStr: fmt.Sprintf(
					"Key must be less than %d characters",
					maxStringLen,
				),
			}
		}
		return request[int]{key: key, action: action}, nil
	case Clear:
		if len(args) < 2 {
			return request[int]{action: ClearAll}, nil
		}
		if len(key) > maxStringLen {
			return request[int]{}, RequestParseError{
				errorStr: fmt.Sprintf(
					"Key must be less than %d characters",
					maxStringLen,
				),
			}
		}
		return request[int]{key: key, action: action}, nil
	}
	panic("Unreachable")
}

func decodeKey(b []byte) string {
	keyLen := int(b[1])
	return string(b[2 : 2+keyLen])
}

func decodeStringData(b []byte) string {
	dataLen := int(b[0])
	return string(b[1 : 1+dataLen])
}

func DecodeRequest(b []byte) Request {
	action := Action(b[0])
	switch action {
	case Store:
		key := decodeKey(b)
		offset := len(key) + 2
		if b[offset] == 0 {
			data := int32(binary.BigEndian.Uint32(b[offset+1:]))
			return request[int]{
				action: action,
				key:    key,
				data:   int(data),
				id:     generateId(),
			}
		}
		data := decodeStringData(b[offset+1:])
		return request[string]{
			action: action,
			key:    key,
			data:   data,
			id:     generateId(),
		}
	case Load, Clear:
		key := decodeKey(b)
		return request[int]{
			action: action,
			key:    key,
			id:     generateId(),
		}
	case ClearAll:
		return request[int]{
			action: action,
			id:     generateId(),
		}
	}
	panic("Unreachable")
}
