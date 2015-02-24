package log

import (
	"bytes"
	"io"
	"os"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

// Logger is the interface that logger.
type Logger interface {
	// Debug calls Logger.Output to print to the logger with DEBUG level.
	// Arguments are handled in the manner of fmt.Print.
	// If the current log level is upper than DEBUG, it won't be the output.
	Debug(v ...interface{})

	// Debugf calls Logger.Output to print to the logger with DEBUG level.
	// Arguments are handled in the manner of fmt.Printf.
	// If the current log level is upper than DEBUG, it won't be the output.
	Debugf(format string, v ...interface{})

	// Debugln calls Logger.Output to print to the logger with DEBUG level.
	// Arguments are handled in the manner of fmt.Println.
	// If the current log level is upper than DEBUG, it won't be the output.
	Debugln(v ...interface{})

	// Info calls Logger.Output to print to the logger with INFO level.
	// Arguments are handled in the manner of fmt.Print.
	// If the current log level is upper than INFO, it won't be the output.
	Info(v ...interface{})

	// Infof calls Logger.Output to print to the logger with INFO level.
	// Arguments are handled in the manner of fmt.Printf.
	// If the current log level is upper than INFO, it won't be the output.
	Infof(format string, v ...interface{})

	// Infoln calls Logger.Output to print to the logger with INFO level.
	// Arguments are handled in the manner of fmt.Println.
	// If the current log level is upper than INFO, it won't be the output.
	Infoln(v ...interface{})

	// Warn calls Logger.Output to print to the logger with WARN level.
	// Arguments are handled in the manner of fmt.Print.
	// If the current log level is upper than WARN, it won't be the output.
	Warn(v ...interface{})

	// Warnf calls Logger.Output to print to the logger with WARN level.
	// Arguments are handled in the manner of fmt.Printf.
	// If the current log level is upper than WARN, it won't be the output.
	Warnf(format string, v ...interface{})

	// Warnln calls Logger.Output to print to the logger with WARN level.
	// Arguments are handled in the manner of fmt.Println.
	// If the current log level is upper than WARN, it won't be the output.
	Warnln(v ...interface{})

	// Error calls Logger.Output to print to the logger with ERROR level.
	// Arguments are handled in the manner of fmt.Print.
	// If the current log level is upper than ERROR, it won't be the output.
	Error(v ...interface{})

	// Errorf calls Logger.Output to print to the logger with ERROR level.
	// Arguments are handled in the manner of fmt.Printf.
	// If the current log level is upper than ERROR, it won't be the output.
	Errorf(format string, v ...interface{})

	// Errorln calls Logger.Output to print to the logger with ERROR level.
	// Arguments are handled in the manner of fmt.Println.
	// If the current log level is upper than ERROR, it won't be the output.
	Errorln(v ...interface{})

	// Fatal calls Logger.Output to print to the logger with FATAL level.
	// Arguments are handled in the manner of fmt.Print.
	// Also calls os.Exit(1) after the output.
	Fatal(v ...interface{})

	// Fatalf calls Logger.Output to print to the logger with FATAL level.
	// Arguments are handled in the manner of fmt.Printf.
	// Also calls os.Exit(1) after the output.
	Fatalf(format string, v ...interface{})

	// Fatalln calls Logger.Output to print to the logger with FATAL level.
	// Arguments are handled in the manner of fmt.Println.
	// Also calls os.Exit(1) after the output.
	Fatalln(v ...interface{})

	// Panic calls Logger.Output to print to the logger with PANIC level.
	// Arguments are handled in the manner of fmt.Print.
	// Also calls panic() after the output.
	Panic(v ...interface{})

	// Panicf calls Logger.Output to print to the logger with PANIC level.
	// Arguments are handled in the manner of fmt.Printf.
	// Also calls panic() after the output.
	Panicf(format string, v ...interface{})

	// Panicln calls Logger.Output to print to the logger with PANIC level.
	// Arguments are handled in the manner of fmt.Println.
	// Also calls panic() after the output.
	Panicln(v ...interface{})

	// Print calls Logger.Output to print to the logger with NONE level.
	// Arguments are handled in the manner of fmt.Print.
	Print(v ...interface{})

	// Printf calls Logger.Output to print to the logger with NONE level.
	// Arguments are handled in the manner of fmt.Printf.
	Printf(format string, v ...interface{})

	// Println calls Logger.Output to print to the logger with NONE level.
	// Arguments are handled in the manner of fmt.Println.
	Println(v ...interface{})

	// Output writes the output for a logging event with the given level.
	// The given message will be format by Formatter. Also a newline is appended
	// to the message before the output.
	Output(level Level, message string)

	// With returns a new Logger with fields.
	With(fields Fields) Logger

	// Level returns the current log level.
	Level() Level

	// SetLevel sets the log level.
	SetLevel(level Level)
}

// New creates a new Logger.
func New(out io.Writer, formatter Formatter, level Level) Logger {
	l := &logger{
		out:         out,
		formatter:   formatter,
		formatFuncs: plainFormats,
		level:       level,
	}
	if w, ok := out.(*os.File); ok && isatty.IsTerminal(w.Fd()) {
		switch w {
		case os.Stdout:
			l.out = colorable.NewColorableStdout()
			l.formatFuncs = coloredFormats
		case os.Stderr:
			l.out = colorable.NewColorableStderr()
			l.formatFuncs = coloredFormats
		}
	}
	return l
}

// logger implements the Logger interface.
type logger struct {
	out         io.Writer
	formatter   Formatter
	formatFuncs [7]formatFunc
	level       Level
	fields      Fields
	buf         bytes.Buffer
	mu          sync.Mutex
}

func (l *logger) Debug(v ...interface{}) {
	if l.Level() <= DEBUG {
		newEntryLogger(l).Debug(v...)
	}
}

func (l *logger) Debugf(format string, v ...interface{}) {
	if l.Level() <= DEBUG {
		newEntryLogger(l).Debugf(format, v...)
	}
}

func (l *logger) Debugln(v ...interface{}) {
	if l.Level() <= DEBUG {
		newEntryLogger(l).Debugln(v...)
	}
}

func (l *logger) Info(v ...interface{}) {
	if l.Level() <= INFO {
		newEntryLogger(l).Info(v...)
	}
}

func (l *logger) Infof(format string, v ...interface{}) {
	if l.Level() <= INFO {
		newEntryLogger(l).Infof(format, v...)
	}
}

func (l *logger) Infoln(v ...interface{}) {
	if l.Level() <= INFO {
		newEntryLogger(l).Infoln(v...)
	}
}

func (l *logger) Warn(v ...interface{}) {
	if l.Level() <= WARN {
		newEntryLogger(l).Warn(v...)
	}
}

func (l *logger) Warnf(format string, v ...interface{}) {
	if l.Level() <= WARN {
		newEntryLogger(l).Warnf(format, v...)
	}
}

func (l *logger) Warnln(v ...interface{}) {
	if l.Level() <= WARN {
		newEntryLogger(l).Warnln(v...)
	}
}

func (l *logger) Error(v ...interface{}) {
	if l.Level() <= ERROR {
		newEntryLogger(l).Error(v...)
	}
}

func (l *logger) Errorf(format string, v ...interface{}) {
	if l.Level() <= ERROR {
		newEntryLogger(l).Errorf(format, v...)
	}
}

func (l *logger) Errorln(v ...interface{}) {
	if l.Level() <= ERROR {
		newEntryLogger(l).Errorln(v...)
	}
}

func (l *logger) Fatal(v ...interface{}) {
	newEntryLogger(l).Fatal(v...)
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	newEntryLogger(l).Fatalf(format, v...)
}

func (l *logger) Fatalln(v ...interface{}) {
	newEntryLogger(l).Fatalln(v...)
}

func (l *logger) Panic(v ...interface{}) {
	newEntryLogger(l).Panic(v...)
}

func (l *logger) Panicf(format string, v ...interface{}) {
	newEntryLogger(l).Panicf(format, v...)
}

func (l *logger) Panicln(v ...interface{}) {
	newEntryLogger(l).Panicln(v...)
}

func (l *logger) Print(v ...interface{}) {
	newEntryLogger(l).Print(v...)
}

func (l *logger) Printf(format string, v ...interface{}) {
	newEntryLogger(l).Printf(format, v...)
}

func (l *logger) Println(v ...interface{}) {
	newEntryLogger(l).Println(v...)
}

func (l *logger) Output(level Level, message string) {
	newEntryLogger(l).Output(level, message)
}

func (l *logger) With(fields Fields) Logger {
	return newEntryLogger(l).With(fields)
}

func (l *logger) Level() Level {
	return Level(atomic.LoadUint32((*uint32)(&l.level)))
}

func (l *logger) SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&l.level), uint32(level))
}

// Level represents a log level.
type Level uint32

// The log levels.
const (
	NONE Level = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	PANIC
)

type formatFunc func(f Formatter, w io.Writer, entry *Entry) error

func makeFormat(esc string) formatFunc {
	return func(f Formatter, w io.Writer, entry *Entry) error {
		io.WriteString(w, esc)
		err := f.Format(w, entry)
		io.WriteString(w, "\x1b[0m")
		return err
	}
}

var coloredFormats = [...]formatFunc{
	Formatter.Format,       // NONE
	Formatter.Format,       // DEBUG
	Formatter.Format,       // INFO
	makeFormat("\x1b[33m"), // WARN
	makeFormat("\x1b[31m"), // ERROR
	makeFormat("\x1b[31m"), // FATAL
	makeFormat("\x1b[31m"), // PANIC
}

var plainFormats = [...]formatFunc{
	Formatter.Format, // NONE
	Formatter.Format, // DEBUG
	Formatter.Format, // INFO
	Formatter.Format, // WARN
	Formatter.Format, // ERROR
	Formatter.Format, // FATAL
	Formatter.Format, // PANIC
}

// Fields represents the key-value pairs in a log Entry.
type Fields map[string]interface{}

// Get returns a value associated with the given key.
func (f Fields) Get(key string) interface{} {
	return f[key]
}

// OrderedKeys returns the keys of f that sorted in increasing order.
// This is used if you need the consistent map iteration order.
// See also http://golang.org/doc/go1.3#map
func (f Fields) OrderedKeys() (keys []string) {
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
