package server

import (
	"io"
	"log"
	"os"

	"github.com/EnemigoPython/go-getit/src/runtime"
)

var logFile *os.File

func configureLogger() {
	if runtime.Config.NoLog {
		return
	}
	logFile, err := os.OpenFile(
		runtime.Config.LogPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	if runtime.Config.Debug {
		multi := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multi)
	} else {
		log.SetOutput(logFile)
	}
	log.Println("Logger initialised")
}

func shutdown() {
	log.Print("Exiting\n\n")
	if !runtime.Config.NoLog {
		logFile.Close()
	}
}
