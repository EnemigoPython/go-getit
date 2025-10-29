package runtime

import (
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
	Action Action
	Key    string
	Data   T
}

type Message interface {
	getDataBytes() []byte
	GetMessageBytes() []byte
}

func (m message[T]) getDataBytes() []byte    { return []byte{} }
func (m message[T]) GetMessageBytes() []byte { return []byte{} }

func ConstructMessage(args []string) (Message, error) {
	action, err := parseAction(args[1])
	if err != nil {
		return message[int]{}, err
	}
	var key string
	var data string
	switch action {
	case Store:
		if len(args) < 3 {
			return message[int]{}, RunTimeParseError{
				RunTimeStr: "need 3 args for store",
			}
		}
		key = args[2]
		data = args[3]
		if i, err := strconv.Atoi(data); err == nil {
			return message[int]{Key: key, Data: i, Action: action}, nil
		}
		return message[string]{Key: key, Data: data, Action: action}, nil
	case Load:
		if len(args) < 2 {
			return message[int]{}, RunTimeParseError{
				RunTimeStr: "need 2 args for load",
			}
		}
		key = args[2]
		return message[int]{Key: key, Action: action}, nil
	case Clear:
	default:
		return message[int]{Action: action}, nil
	}
	return message[int]{}, RunTimeParseError{
		RunTimeStr: "invalid ",
	}
}
