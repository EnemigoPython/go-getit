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

func (e RunTimeParseError) Error() string {
	return fmt.Sprintf("Error initialising runtime; invalid runtime: %s", e.RunTimeStr)
}

func (r RunTime) String() string {
	return [...]string{"Server", "Client"}[r]
}

func (r RunTime) ToLower() string {
	return [...]string{"server", "client"}[r]
}

type RunTimeParseError struct {
	RunTimeStr string
}

func parseRunTime(s string) (RunTime, error) {
	switch strings.ToLower(s) {
	case "":
		return RunTime(0), RunTimeParseError{RunTimeStr: "<empty>"}
	case Server.ToLower():
		return Server, nil
	case Client.ToLower():
		return Client, nil
	default:
		return RunTime(0), RunTimeParseError{RunTimeStr: s}
	}
}

type _Config struct {
	RunTime   RunTime
	Port      int
	StoreName string
}

var Config _Config

func SocketAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", Config.Port)
}

func ParseConfig(runTimeStr string, port int, storeName string) (_Config, error) {
	runTime, err := parseRunTime(runTimeStr)
	if err != nil {
		return _Config{}, err
	}
	Config = _Config{
		RunTime:   runTime,
		Port:      port,
		StoreName: storeName,
	}
	return Config, nil
}
