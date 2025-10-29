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

type Message[T intOrString] struct {
	Action Action
	Key    string
	Data   T
}

type MessageInterface interface {
	GetDataBytes() []byte
}

func (m Message[T]) GetDataBytes() []byte { return []byte{} }

func ConstructMessage(args []string) (MessageInterface, error) {
	action, err := parseAction(args[1])
	if err != nil {
		return Message[int]{}, err
	}
	var key string
	var data string
	switch action {
	case Store:
		if len(args) < 3 {
			return Message[int]{}, RunTimeParseError{
				RunTimeStr: "need 3 args for store",
			}
		}
		key = args[2]
		data = args[3]
		if i, err := strconv.Atoi(data); err == nil {
			return Message[int]{Key: key, Data: i}, nil
		}
		return Message[string]{Key: key, Data: data}, nil
	case Load:
		if len(args) < 2 {
			return Message[int]{}, RunTimeParseError{
				RunTimeStr: "need 2 args for load",
			}
		}
	case Clear:
	}
	return Message[int]{}, RunTimeParseError{
		RunTimeStr: "invalid ",
	}
}
