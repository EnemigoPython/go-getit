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
)

func (a Action) String() string {
	return [...]string{"Store", "Load", "Clear"}[a]
}

func (a Action) ToLower() string {
	return [...]string{"store", "load", "clear"}[a]
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

type request[T types.IntOrString] struct {
	action Action
	key    string
	data   T
}

type Request interface {
	EncodeRequest() []byte
	GetAction() Action
}

func (r request[T]) writeKeyBytes(buf *bytes.Buffer) {
	buf.WriteByte(byte(len(r.key))) // number of bytes
	buf.Write([]byte(r.key))
}

func (r request[T]) writeDataBytes(buf *bytes.Buffer) {
	switch d := any(r.data).(type) {
	case int:
		buf.WriteByte(byte(0)) // type of data: int
		binary.Write(buf, binary.BigEndian, uint16(d))
	case string:
		buf.WriteByte(byte(1))      // type of data: string
		buf.WriteByte(byte(len(d))) // number of bytes
		buf.Write([]byte(d))
	}
}

func (r request[T]) EncodeRequest() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(r.action))
	switch r.action {
	case Store:
		r.writeKeyBytes(buf)
		r.writeDataBytes(buf)
	case Load:
		r.writeKeyBytes(buf)
	case Clear:
		// no extra data fields needed
	}
	return buf.Bytes()
}

func (r request[T]) GetAction() Action { return r.action }

func (r request[T]) String() string {
	var body string
	switch r.action {
	case Store:
		switch d := any(r.data).(type) {
		case int:
			body = fmt.Sprintf("%s [%s: %d]", r.action, r.key, d)
		case string:
			body = fmt.Sprintf("%s [%s: '%s']", r.action, r.key, d)
		default:
			panic("Unreachable")
		}
	case Load:
		body = fmt.Sprintf("%s [%s]", r.action, r.key)
	case Clear:
		body = r.action.String()
	default:
		panic("Unreachable")
	}
	return fmt.Sprintf("Request<%s>", body)
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
		data = args[2]
		if i, err := strconv.Atoi(data); err == nil {
			if i < 0 || i > math.MaxUint16 {
				return request[int]{}, RequestParseError{
					errorStr: fmt.Sprintf(
						"invalid int data (must be 0-%d)",
						math.MaxUint16,
					),
				}
			}
			return request[int]{key: key, data: i, action: action}, nil
		}
		return request[string]{key: key, data: data, action: action}, nil
	case Load:
		if len(args) < 1 {
			return request[int]{}, RequestParseError{
				errorStr: "need 2 args for load",
			}
		}
		key = args[1]
		return request[int]{key: key, action: action}, nil
	case Clear:
		return request[int]{action: action}, nil
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
			data := int(binary.BigEndian.Uint16(b[offset+1:]))
			return request[int]{
				action: action,
				key:    key,
				data:   data,
			}
		}
		data := decodeStringData(b[offset+1:])
		return request[string]{
			action: action,
			key:    key,
			data:   data,
		}
	case Load:
		key := decodeKey(b)
		return request[int]{
			action: action,
			key:    key,
		}
	case Clear:
		return request[int]{
			action: action,
		}
	}
	panic("Unreachable")
}
