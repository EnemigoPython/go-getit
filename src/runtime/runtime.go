package runtime

import (
	"fmt"
	"os"
	"path/filepath"
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
	NoLog     bool
	StorePath string
	TempPath  string
	LogPath   string
}

var Config _Config

func SocketAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", Config.Port)
}

func getStorePath(absDir string, storeName string) string {
	storePath := filepath.Join(absDir, fmt.Sprintf("%s.bin", storeName))
	return storePath
}

func getTempPath(absDir string, storeName string) string {
	tempPath := filepath.Join(absDir, fmt.Sprintf("%s.temp.bin", storeName))
	return tempPath
}

func getLogPath(absDir string, storeName string, debug bool) string {
	var logName string
	if debug {
		logName = fmt.Sprintf("%s.debug.log", storeName)
	} else {
		logName = fmt.Sprintf("%s.log", storeName)
	}
	logPath := filepath.Join(absDir, logName)
	return logPath
}

func ParseConfig(
	runTimeStr string,
	port int,
	storeName string,
	debug bool,
	noLog bool,
) (_Config, error) {
	runTime, err := parseRunTime(runTimeStr)
	if err != nil {
		return _Config{}, err
	}
	absPath, err := os.Executable()
	if err != nil {
		return _Config{}, err
	}
	absDir := filepath.Dir(absPath)
	Config = _Config{
		RunTime:   runTime,
		Port:      port,
		StoreName: storeName,
		Debug:     debug,
		NoLog:     noLog,
		StorePath: getStorePath(absDir, storeName),
		TempPath:  getTempPath(absDir, storeName),
		LogPath:   getLogPath(absDir, storeName, debug),
	}
	return Config, nil
}
