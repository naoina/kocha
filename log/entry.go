package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/naoina/kocha/util"
)

// Entry represents a log entry.
type Entry struct {
	Level   Level     // log level.
	Time    time.Time // time of the log event.
	Message string    // log message (optional).
	Fields  Fields    // extra fields of the log entry (optional).
}

// entryLogger implements the Logger interface.
type entryLogger struct {
	entry  *Entry
	logger *logger
	mu     sync.Mutex
}

func newEntryLogger(logger *logger) *entryLogger {
	return &entryLogger{
		logger: logger,
		entry:  &Entry{},
	}
}

func (l *entryLogger) Debug(v ...interface{}) {
	if l.logger.Level() <= DEBUG {
		l.Output(DEBUG, fmt.Sprint(v...))
	}
}

func (l *entryLogger) Debugf(format string, v ...interface{}) {
	if l.logger.Level() <= DEBUG {
		l.Output(DEBUG, fmt.Sprintf(format, v...))
	}
}

func (l *entryLogger) Debugln(v ...interface{}) {
	if l.logger.Level() <= DEBUG {
		l.Output(DEBUG, fmt.Sprint(v...))
	}
}

func (l *entryLogger) Info(v ...interface{}) {
	if l.logger.Level() <= INFO {
		l.Output(INFO, fmt.Sprint(v...))
	}
}

func (l *entryLogger) Infof(format string, v ...interface{}) {
	if l.logger.Level() <= INFO {
		l.Output(INFO, fmt.Sprintf(format, v...))
	}
}

func (l *entryLogger) Infoln(v ...interface{}) {
	if l.logger.Level() <= INFO {
		l.Output(INFO, fmt.Sprint(v...))
	}
}

func (l *entryLogger) Warn(v ...interface{}) {
	if l.logger.Level() <= WARN {
		l.Output(WARN, fmt.Sprint(v...))
	}
}

func (l *entryLogger) Warnf(format string, v ...interface{}) {
	if l.logger.Level() <= WARN {
		l.Output(WARN, fmt.Sprintf(format, v...))
	}
}

func (l *entryLogger) Warnln(v ...interface{}) {
	if l.logger.Level() <= WARN {
		l.Output(WARN, fmt.Sprint(v...))
	}
}

func (l *entryLogger) Error(v ...interface{}) {
	if l.logger.Level() <= ERROR {
		l.Output(ERROR, fmt.Sprint(v...))
	}
}

func (l *entryLogger) Errorf(format string, v ...interface{}) {
	if l.logger.Level() <= ERROR {
		l.Output(ERROR, fmt.Sprintf(format, v...))
	}
}

func (l *entryLogger) Errorln(v ...interface{}) {
	if l.logger.Level() <= ERROR {
		l.Output(ERROR, fmt.Sprint(v...))
	}
}

func (l *entryLogger) Fatal(v ...interface{}) {
	l.Output(FATAL, fmt.Sprint(v...))
	os.Exit(1)
}

func (l *entryLogger) Fatalf(format string, v ...interface{}) {
	l.Output(FATAL, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l *entryLogger) Fatalln(v ...interface{}) {
	l.Output(FATAL, fmt.Sprint(v...))
	os.Exit(1)
}

func (l *entryLogger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Output(PANIC, s)
	panic(s)
}

func (l *entryLogger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Output(PANIC, s)
	panic(s)
}

func (l *entryLogger) Panicln(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Output(PANIC, s)
	panic(s)
}

func (l *entryLogger) Print(v ...interface{}) {
	l.Output(NONE, fmt.Sprint(v...))
}

func (l *entryLogger) Printf(format string, v ...interface{}) {
	l.Output(NONE, fmt.Sprintf(format, v...))
}

func (l *entryLogger) Println(v ...interface{}) {
	l.Output(NONE, fmt.Sprint(v...))
}

func (l *entryLogger) Output(level Level, message string) {
	l.logger.mu.Lock()
	defer l.logger.mu.Unlock()
	l.entry.Level = level
	l.entry.Time = util.Now()
	l.entry.Message = message
	l.logger.buf.Reset()
	if err := l.logger.formatter.Format(&l.logger.buf, l.entry); err != nil {
		fmt.Fprintf(os.Stderr, "kocha: log: %v\n", err)
	}
	l.logger.buf.WriteByte('\n')
	if _, err := io.Copy(l.logger.out, &l.logger.buf); err != nil {
		fmt.Fprintf(os.Stderr, "kocha: log: failed to write log: %v\n", err)
	}
}

func (l *entryLogger) With(fields Fields) Logger {
	l.mu.Lock()
	l.entry.Fields = fields
	l.mu.Unlock()
	return l
}

func (l *entryLogger) Level() Level {
	return l.logger.Level()
}

func (l *entryLogger) SetLevel(level Level) {
	l.logger.SetLevel(level)
}
