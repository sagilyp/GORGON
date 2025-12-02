package logger

import (
	"log"
	"os"
	"sync"
	"fmt"
)

var (
	logFilePath = "/var/log/gorgon/fast.log"
	fileLogger *log.Logger
	once sync.Once
)

type Logger interface {
	Logf(format string, v ...interface{})
}

// LoggerImpl реализует интерфейс Logger для DI
type LoggerImpl struct{}

func (l *LoggerImpl) Logf(format string, v ...interface{}) {
    Logf(format, v...)
}

func initLogger() {
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("logger: failed to open log file: %v", err)
		fileLogger = nil
		return
	}
	fileLogger = log.New(f, "", log.LstdFlags)
}

func Logf(format string, v ...interface{}) {
	once.Do(initLogger)
	msg := format
	if len(v) > 0 {
		msg = fmt.Sprintf(format, v...)
	}
	log.Printf(msg) // стандартный вывод (journalctl)
	if fileLogger != nil {
		fileLogger.Printf(msg)
	}
}
