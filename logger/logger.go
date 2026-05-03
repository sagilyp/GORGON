package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var (
	logFilePath = "/var/log/gorgon/fast.log"
	fileLogger *log.Logger
	once sync.Once
	hostname string
)

type Logger interface {
	Logf(format string, v ...interface{})
}

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
	fileLogger = log.New(f, "", 0)
	host,err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	} 
	hostname = host
}

func formatRFC5424(msg string) string {
	pri := "<134>1"
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	appName := "GORGON_IDS"
	procID := fmt.Sprintf("%d", os.Getpid())
	msgID := "-"
	structData := "-"
	return fmt.Sprintf("%s %s %s %s %s %s %s %s", pri, timestamp, hostname, appName, procID, msgID, structData, msg)
}

func Logf(format string, v ...interface{}) {
	once.Do(initLogger)
	rawMsg := format
	if len(v) > 0 {
		rawMsg = fmt.Sprintf(format, v...)
	}
	rfcMsg := formatRFC5424(rawMsg)
	log.Printf(rfcMsg) // стандартный вывод journalctl
	if fileLogger != nil {
		fileLogger.Println(rfcMsg)
	}
}
