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

type Action int

const (
	Store Action = iota
	Load
	Clear
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

type _Config struct {
	RunTime RunTime
	Port    int
}

func SocketAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", Config.Port)
}

var Config _Config

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
