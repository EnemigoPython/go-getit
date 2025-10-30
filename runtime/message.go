package runtime

import (
	"fmt"
	"strconv"
	"strings"
)

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
		return Action(0), RunTimeParseError{RunTimeStr: "<empty>"}
	case Store.ToLower():
		return Store, nil
	case Load.ToLower():
		return Load, nil
	case Clear.ToLower():
		return Clear, nil
	default:
		return Action(0), RunTimeParseError{RunTimeStr: s}
	}
}

type intOrString interface {
	~int | ~string
}

type message[T intOrString] struct {
	action Action
	key    string
	data   T
}

type Message interface {
	getDataBytes() []byte
	GetMessageBytes() []byte
}

func (m message[T]) getDataBytes() []byte { return []byte{} }

func (m message[T]) GetMessageBytes() []byte { return []byte("Hi!") }

func (m message[T]) String() string {
	switch m.action {
	case Store:
		switch d := any(m.data).(type) {
		case int:
			return fmt.Sprintf("%s: [%s=%d]", m.action, m.key, d)
		case string:
			return fmt.Sprintf("%s: [%s=%s]", m.action, m.key, d)
		default:
			return fmt.Sprintf("%s: [%s=%d]", m.action, m.key, d)
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
			return message[int]{}, RunTimeParseError{
				RunTimeStr: "need 3 args for store",
			}
		}
		key = args[1]
		data = args[2]
		if i, err := strconv.Atoi(data); err == nil {
			return message[int]{key: key, data: i, action: action}, nil
		}
		return message[string]{key: key, data: data, action: action}, nil
	case Load:
		if len(args) < 1 {
			return message[int]{}, RunTimeParseError{
				RunTimeStr: "need 2 args for load",
			}
		}
		key = args[1]
		return message[int]{key: key, action: action}, nil
	case Clear:
		return message[int]{action: action}, nil
	default:
		return message[int]{}, RunTimeParseError{
			RunTimeStr: "invalid action",
		}
	}
}
