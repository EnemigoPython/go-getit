package server

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/EnemigoPython/go-getit/runtime"
)

var logFile *os.File

func configureLogger() {
	var logName string
	if runtime.Config.Debug {
		logName = fmt.Sprintf("%s.debug.log", runtime.Config.StoreName)
	} else {
		logName = fmt.Sprintf("%s.log", runtime.Config.StoreName)
	}
	logFile, err := os.OpenFile(
		logName,
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
	logFile.Close()
}
