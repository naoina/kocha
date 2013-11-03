package kocha

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	defaultLflag = log.Ldate | log.Ltime
)

var (
	nullLogger = log.New(ioutil.Discard, "", 0)
)

func initLogger(logger *Logger) *Logger {
	if logger == nil {
		logger = &Logger{}
	}
	if logger.DEBUG == nil {
		logger.DEBUG = Loggers{nullLogger}
	}
	if logger.INFO == nil {
		logger.INFO = Loggers{nullLogger}
	}
	if logger.WARN == nil {
		logger.WARN = Loggers{nullLogger}
	}
	if logger.ERROR == nil {
		logger.ERROR = Loggers{nullLogger}
	}
	setPrefix := func(loggers Loggers, prefix string) {
		for _, logger := range loggers {
			logger.SetPrefix(prefix)
		}
	}
	setPrefix(logger.DEBUG, "[DEBUG] ")
	setPrefix(logger.INFO, "[INFO]  ")
	setPrefix(logger.WARN, "[WARN]  ")
	setPrefix(logger.ERROR, "[ERROR] ")
	return logger
}

func NullLogger() *log.Logger {
	return nullLogger
}

func ConsoleLogger(flag int) *log.Logger {
	if flag == -1 {
		flag = defaultLflag
	}
	return log.New(os.Stdout, "", flag)
}

type Loggers []*log.Logger

type Logger struct {
	DEBUG Loggers
	INFO  Loggers
	WARN  Loggers
	ERROR Loggers
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.output(l.DEBUG, format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.output(l.INFO, format, v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.output(l.WARN, format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.output(l.ERROR, format, v...)
}

func (l *Logger) output(loggers Loggers, format string, v ...interface{}) {
	output := fmt.Sprintf(format+"\n", v...)
	for _, logger := range loggers {
		logger.Output(2, output)
	}
}
