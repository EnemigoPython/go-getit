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

type RunTimeParseError struct {
	runTimeStr string
}

func (e RunTimeParseError) Error() string {
	return fmt.Sprintf("Error initialising runtime; invalid runtime: %s", e.runTimeStr)
}

func (r RunTime) String() string {
	return [...]string{"Server", "Client"}[r]
}

func (r RunTime) ToLower() string {
	return [...]string{"server", "client"}[r]
}

func parseRunTime(s string) (RunTime, error) {
	switch strings.ToLower(s) {
	case "":
		return RunTime(0), RunTimeParseError{runTimeStr: "<empty>"}
	case Server.ToLower():
		return Server, nil
	case Client.ToLower():
		return Client, nil
	default:
		return RunTime(0), RunTimeParseError{runTimeStr: s}
	}
}

type _Config struct {
	RunTime   RunTime
	Port      int
	StoreName string
	Debug     bool
}

var Config _Config

func SocketAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", Config.Port)
}

func ParseConfig(
	runTimeStr string,
	port int,
	storeName string,
	debug bool,
) (_Config, error) {
	runTime, err := parseRunTime(runTimeStr)
	if err != nil {
		return _Config{}, err
	}
	Config = _Config{
		RunTime:   runTime,
		Port:      port,
		StoreName: storeName,
		Debug:     debug,
	}
	return Config, nil
}
