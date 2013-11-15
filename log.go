package kocha

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultLflag = log.Ldate | log.Ltime
)

func initLogger(logger *Logger) *Logger {
	if logger == nil {
		logger = &Logger{}
	}
	if logger.DEBUG == nil {
		logger.DEBUG = Loggers{NullLogger()}
	}
	if logger.INFO == nil {
		logger.INFO = Loggers{NullLogger()}
	}
	if logger.WARN == nil {
		logger.WARN = Loggers{NullLogger()}
	}
	if logger.ERROR == nil {
		logger.ERROR = Loggers{NullLogger()}
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

type logger interface {
	Output(calldepth int, s string) error
	SetPrefix(prefix string)
	GoString() string
}

type nullLogger struct {
	*log.Logger
}

func (l *nullLogger) GoString() string {
	return "kocha.NullLogger()"
}

func NullLogger() logger {
	return &nullLogger{log.New(ioutil.Discard, "", 0)}
}

type consoleLogger struct {
	*log.Logger
}

func (l *consoleLogger) GoString() string {
	return fmt.Sprintf("kocha.ConsoleLogger(%d)", l.Flags())
}

func ConsoleLogger(flag int) logger {
	if flag == -1 {
		flag = defaultLflag
	}
	return &consoleLogger{log.New(os.Stdout, "", flag)}
}

type fileLogger struct {
	*log.Logger
	path string
}

func (l *fileLogger) GoString() string {
	return fmt.Sprintf("kocha.FileLogger(%q, %d)", l.path, l.Flags())
}

func FileLogger(path string, flag int) logger {
	if flag == -1 {
		flag = defaultLflag
	}
	path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		panic(err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return &fileLogger{log.New(file, "", flag), path}
}

type Loggers []logger

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
