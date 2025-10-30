package runtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type MessageParseError struct {
	errorStr string
}

func (e MessageParseError) Error() string {
	return fmt.Sprintf("Error parsing message; %s", e.errorStr)
}

type Action int

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
		return Action(0), MessageParseError{errorStr: "invalid action: <empty>"}
	case Store.ToLower():
		return Store, nil
	case Load.ToLower():
		return Load, nil
	case Clear.ToLower():
		return Clear, nil
	default:
		return Action(0), MessageParseError{errorStr: s}
	}
}

type intOrString interface {
	int | string
}

type message[T intOrString] struct {
	action Action
	key    string
	data   T
}

type Message interface {
	EncodeMessage() []byte
}

func (m message[T]) writeKeyBytes(buf *bytes.Buffer) {
	buf.WriteByte(byte(len(m.key))) // number of bytes
	buf.Write([]byte(m.key))
}

func (m message[T]) writeDataBytes(buf *bytes.Buffer) {
	switch d := any(m.data).(type) {
	case int:
		buf.WriteByte(byte(0)) // type of data: int
		binary.Write(buf, binary.BigEndian, uint16(d))
	case string:
		buf.WriteByte(byte(1))      // type of data: string
		buf.WriteByte(byte(len(d))) // number of bytes
		buf.Write([]byte(d))
	}
}

func (m message[T]) EncodeMessage() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(m.action))
	switch m.action {
	case Store:
		m.writeKeyBytes(buf)
		m.writeDataBytes(buf)
	case Load:
		m.writeKeyBytes(buf)
	case Clear:
		// no extra data fields needed
	}
	return buf.Bytes()
}

func (m message[T]) String() string {
	switch m.action {
	case Store:
		switch d := any(m.data).(type) {
		case int:
			return fmt.Sprintf("%s: [%s=%d]", m.action, m.key, d)
		case string:
			return fmt.Sprintf("%s: [%s=%s]", m.action, m.key, d)
		default:
			panic("Unreachable")
		}
	case Load:
		return fmt.Sprintf("%s: [%s]", m.action, m.key)
	case Clear:
		return m.action.String()
	default:
		return ""
	}
}

func ConstructMessage(args []string) (Message, error) {
	action, err := parseAction(args[0])
	if err != nil {
		return message[int]{}, err
	}
	var key string
	var data string
	switch action {
	case Store:
		if len(args) < 2 {
			return message[int]{}, MessageParseError{
				errorStr: "need 3 args for store",
			}
		}
		key = args[1]
		data = args[2]
		if i, err := strconv.Atoi(data); err == nil {
			if i < 0 || i > math.MaxUint16 {
				return message[int]{}, MessageParseError{
					errorStr: fmt.Sprintf(
						"invalid int data (must be 0-%d)",
						math.MaxUint16,
					),
				}
			}
			return message[int]{key: key, data: i, action: action}, nil
		}
		return message[string]{key: key, data: data, action: action}, nil
	case Load:
		if len(args) < 1 {
			return message[int]{}, MessageParseError{
				errorStr: "need 2 args for load",
			}
		}
		key = args[1]
		return message[int]{key: key, action: action}, nil
	case Clear:
		return message[int]{action: action}, nil
	}
	panic("Unreachable")
}

func decodeKey(b []byte) string {
	keyLen := int(b[1])
	return string(b[2 : 2+keyLen])
}

func decodeData(b []byte, offset int) int {
	return 0
}

func DecodeMessage(b []byte) Message {
	action := Action(b[0])
	switch action {
	case Store:
		key := decodeKey(b)
		offset := len(key) + 3
		data := decodeData(b, offset)
		return message[int]{
			action: action,
			key:    key,
			data:   data,
		}
	case Load:
		key := decodeKey(b)
		return message[int]{
			action: action,
			key:    key,
		}
	case Clear:
		return message[int]{
			action: action,
		}
	}
	panic("Unreachable")
}
