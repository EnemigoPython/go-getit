package runtime

import (
	"fmt"
	"strings"
)

type RunTime int

const (
	Server RunTime = iota
	Client
)

func (r RunTime) String() string {
	return [...]string{"Server", "Client"}[r]
}

func (r RunTime) LowerString() string {
	return [...]string{"server", "client"}[r]
}

type RunTimeError struct {
	RunTimeStr string
}

func (e RunTimeError) Error() string {
	return fmt.Sprintf("Error initialising runtime; invalid runtime: %s", e.RunTimeStr)
}

type Action int

const (
	Store Action = iota
	Load
	Clear
)

type IntOrString interface {
	~int | ~string
}

type Message[T IntOrString] struct {
	Action Action
	Data   T
}

type _Config struct {
	RunTime RunTime
	Port    int
}

var Config _Config

func SocketAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", Config.Port)
}

func parseRunTime(s string) (RunTime, error) {
	switch strings.ToLower(s) {
	case "":
		return RunTime(0), RunTimeError{RunTimeStr: "<empty>"}
	case Server.LowerString():
		return Server, nil
	case Client.LowerString():
		return Client, nil
	default:
		return RunTime(0), RunTimeError{RunTimeStr: s}
	}
}

func ParseConfig(runTimeStr string, port int) (_Config, error) {
	runTime, err := parseRunTime(runTimeStr)
	if err != nil {
		return _Config{}, err
	}
	Config = _Config{
		RunTime: runTime,
		Port:    port,
	}
	return Config, nil
}
