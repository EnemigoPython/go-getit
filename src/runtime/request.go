package runtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/EnemigoPython/go-getit/src/types"
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
	Add
	Sub
	Load
	Clear
	ClearAll
	Keys
	Values
	Items
	Resize
	Count
	Size
	Space
	Exit
)

type ArithmeticType int

const (
	A_Add ArithmeticType = iota
	A_Sub
)

func (a Action) String() string {
	return [...]string{
		"Store",
		"Add",
		"Sub",
		"Load",
		"Clear",
		"ClearAll",
		"Keys",
		"Values",
		"Items",
		"Resize",
		"Count",
		"Size",
		"Space",
		"Exit",
	}[a]
}

func (a Action) ToLower() string {
	return [...]string{
		"store",
		"add",
		"sub",
		"load",
		"clear",
		"clearall",
		"keys",
		"values",
		"items",
		"resize",
		"count",
		"size",
		"space",
		"exit",
	}[a]
}

func parseAction(s string) (Action, error) {
	switch strings.ToLower(s) {
	case "":
		return Action(0), RequestParseError{errorStr: "invalid action: <empty>"}
	case Store.ToLower():
		return Store, nil
	case Add.ToLower():
		return Add, nil
	case Sub.ToLower():
		return Sub, nil
	case Load.ToLower():
		return Load, nil
	case Clear.ToLower():
		return Clear, nil
	case Keys.ToLower():
		return Keys, nil
	case Values.ToLower():
		return Values, nil
	case Items.ToLower():
		return Items, nil
	case Resize.ToLower():
		return Resize, nil
	case Count.ToLower():
		return Count, nil
	case Size.ToLower():
		return Size, nil
	case Space.ToLower():
		return Space, nil
	case Exit.ToLower():
		return Exit, nil
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
	action   Action
	key      string
	data     T
	id       uint8
	internal bool
}

type Request interface {
	GetAction() Action
	GetKey() string
	GetId() uint8
	GetIntData() (int, error)
	GetStringData() (string, error)
	IsStream() bool
	HasData() bool
	ArithmeticOperation(ArithmeticType, int) (int, error)
	EncodeRequest() []byte
	EncodeFileBytes() []byte
}

func (r request[T]) GetAction() Action { return r.action }
func (r request[T]) GetKey() string    { return r.key }
func (r request[T]) GetId() uint8      { return r.id }

func (r request[T]) GetIntData() (int, error) {
	switch d := any(r.data).(type) {
	case int:
		return d, nil
	case string:
		return 0, RequestParseError{errorStr: "Wrong data payload"}
	}
	panic("Unreachable")
}

func (r request[T]) GetStringData() (string, error) {
	switch d := any(r.data).(type) {
	case string:
		return d, nil
	case int:
		return "", RequestParseError{errorStr: "Wrong data payload"}
	}
	panic("Unreachable")
}

func (r request[T]) IsStream() bool {
	switch r.action {
	case Keys, Values, Items:
		return true
	default:
		return false
	}
}

func (r request[T]) HasData() bool {
	switch r.action {
	case
		Store,
		Add,
		Sub,
		Load,
		Keys,
		Values,
		Items,
		Count,
		Size,
		Space:
		return true
	default:
		return false
	}
}

// Perform arithmetic on the request where i is the current stored value
func (r request[T]) ArithmeticOperation(a ArithmeticType, i int) (int, error) {
	switch d := any(r.data).(type) {
	case int:
		switch a {
		case A_Add:
			return i + d, nil
		case A_Sub:
			return i - d, nil
		}
	case string:
		return 0, RequestParseError{errorStr: "string is invalid data type"}
	}
	return 0, nil
}

func (r request[T]) writeKeyBytes(buf *bytes.Buffer, pad bool) {
	WriteKeyBytes(buf, r.key, pad)
}

func (r request[T]) writeDataBytes(buf *bytes.Buffer, pad bool) {
	switch d := any(r.data).(type) {
	case int:
		WriteIntBytes(buf, d, pad)
	case string:
		WriteStringBytes(buf, d, pad)
	}
}

func (r request[T]) EncodeRequest() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(r.action))
	switch r.action {
	case Store, Add, Sub:
		r.writeKeyBytes(buf, false)
		r.writeDataBytes(buf, false)
	case Load, Clear, Space:
		r.writeKeyBytes(buf, false)
	case Resize:
		r.writeDataBytes(buf, false)
	default:
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
	case Store, Add, Sub:
		switch d := any(r.data).(type) {
		case int:
			body = fmt.Sprintf("%s[%s:%d]", r.action, r.key, d)
		case string:
			body = fmt.Sprintf("%s[%s:'%s']", r.action, r.key, d)
		default:
			panic("Unreachable")
		}
	case Resize:
		switch d := any(r.data).(type) {
		case int:
			if r.internal {
				body = fmt.Sprintf("Auto%s[%d]", r.action, d)
			} else {
				body = fmt.Sprintf("%s[%d]", r.action, d)
			}
		default:
			panic("Unreachable")
		}
	case Load, Clear, Space:
		body = fmt.Sprintf("%s[%s]", r.action, r.key)
	default:
		body = r.action.String()
	}
	return fmt.Sprintf("Request(%d)<%s>", r.id, body)
}

func ConstructRequest(args []string, internal bool) (Request, error) {
	if len(args) == 0 {
		return request[int]{}, RequestParseError{
			errorStr: "enter a command",
		}
	}
	action, err := parseAction(args[0])
	if err != nil {
		return request[int]{}, err
	}
	var key string
	var data string
	switch a := action; a {
	case Store:
		if len(args) < 3 {
			return request[int]{}, RequestParseError{
				errorStr: "need 3 args for store",
			}
		}
		key = args[1]
		if len(key) > maxStringLen {
			return request[int]{}, RequestParseError{
				errorStr: fmt.Sprintf(
					"key must be less than %d characters",
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
			return request[int]{
				key:      key,
				data:     i,
				action:   action,
				internal: internal,
				id:       generateId(),
			}, nil
		}
		if len(data) > maxStringLen {
			return request[int]{}, RequestParseError{
				errorStr: fmt.Sprintf(
					"data must be less than %d characters",
					maxStringLen,
				),
			}
		}
		return request[string]{
			key:      key,
			data:     data,
			action:   action,
			internal: internal,
			id:       generateId(),
		}, nil
	case Resize:
		if len(args) < 2 {
			return request[int]{}, RequestParseError{
				errorStr: "need 2 args for resize",
			}
		}
		data = args[1]
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
			return request[int]{
				data:     i,
				action:   action,
				internal: internal,
				id:       generateId(),
			}, nil
		}
		return request[int]{}, RequestParseError{
			errorStr: "data for resize must be an integer",
		}
	case Add, Sub:
		if len(args) < 3 {
			return request[int]{}, RequestParseError{
				errorStr: fmt.Sprintf(
					"need 3 args for %s",
					a.ToLower(),
				),
			}
		}
		key = args[1]
		if len(key) > maxStringLen {
			return request[int]{}, RequestParseError{
				errorStr: fmt.Sprintf(
					"key must be less than %d characters",
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
			return request[int]{
				key:      key,
				data:     i,
				action:   action,
				internal: internal,
				id:       generateId(),
			}, nil
		}
		return request[int]{}, RequestParseError{
			errorStr: fmt.Sprintf(
				"data for %s must be an integer",
				a.ToLower(),
			),
		}
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
		return request[int]{
			key:      key,
			action:   action,
			internal: internal,
			id:       generateId(),
		}, nil
	case Clear:
		if len(args) < 2 {
			return request[int]{action: ClearAll, internal: internal}, nil
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
	case Space:
		if len(args) < 2 {
			return request[int]{
				key:      "current",
				action:   action,
				internal: internal,
				id:       generateId(),
			}, nil
		}
		key = args[1]
		switch strings.ToLower(key) {
		case "c", "current":
			return request[int]{
				key:      "current",
				action:   action,
				internal: internal,
				id:       generateId(),
			}, nil
		case "e", "empty":
			return request[int]{
				key:      "empty",
				action:   action,
				internal: internal,
				id:       generateId(),
			}, nil
		default:
			return request[int]{}, RequestParseError{
				"Space verb must be 'current' or 'empty'",
			}
		}
	default:
		return request[int]{
			action:   action,
			internal: internal,
			id:       generateId(),
		}, nil
	}
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
	case Store, Add, Sub:
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
	case Load, Clear, Space:
		key := decodeKey(b)
		return request[int]{
			action: action,
			key:    key,
			id:     generateId(),
		}
	case Resize:
		data := int32(binary.BigEndian.Uint32(b[2:]))
		return request[int]{
			action: action,
			data:   int(data),
			id:     generateId(),
		}
	default:
		return request[int]{
			action: action,
			id:     generateId(),
		}
	}
}
