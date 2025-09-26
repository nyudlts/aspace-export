package aspace_xport

import (
	"fmt"
	"log"
	"os"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
)

var (
	debug   = false
	Logfile string
	logger  *os.File
)

func getLogLevelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "[DEBUG]"
	case INFO:
		return "[INFO]"
	case WARNING:
		return "[WARNING]"
	case ERROR:
		return "[ERROR]"
	case FATAL:
		return "[FATAL]"
	default:
		panic(fmt.Errorf("log level %v is not supported", level))
	}
}

func CreateLogger(dbug bool, logfileName string) error {

	//create a log file
	Logfile = logfileName

	var err error
	logger, err = os.Create(Logfile)
	if err != nil {
		return err
	}

	//set the logger output to the log file
	log.SetPrefix("aspace-export")

	log.SetOutput(logger)
	PrintAndLog(fmt.Sprintf("logging to %s", Logfile), INFO)
	debug = dbug
	return nil
}

func CloseLogger() error {
	err := logger.Close()
	if err != nil {
		return err
	}
	return nil
}

// logging and printing functions
func PrintAndLog(msg string, logLevel LogLevel) {
	if logLevel == DEBUG && debug == false {

	} else {
		level := getLogLevelString(logLevel)
		fmt.Printf("%s %s\n", level, msg)
		log.Printf("%s %s", level, msg)
	}
}

func PrintOnly(msg string, logLevel LogLevel) {
	if logLevel == DEBUG && debug == false {

	} else {
		level := getLogLevelString(logLevel)
		fmt.Printf("%s %s\n", level, msg)
	}
}

func LogOnly(msg string, logLevel LogLevel) {
	if logLevel == DEBUG && debug == false {

	} else {
		level := getLogLevelString(logLevel)
		log.Printf("%s %s\n", level, msg)
	}
}
