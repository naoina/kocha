package log_test

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/naoina/kocha/log"
	"github.com/naoina/kocha/util"
)

func TestLogger_Debug(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:DEBUG\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, ""},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Debug(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Debug(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Debugf(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	expected := "level:DEBUG\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, ""},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Debugf(format, msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Debugf(%#v, %#v) prints %#v; want %#v`, v.level, format, msg, actual, expected)
		}
	}
}

func TestLogger_Debugln(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:DEBUG\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, ""},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Debugln(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Debugln(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Info(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:INFO\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Info(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Info(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Infof(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	expected := "level:INFO\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Infof(format, msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Infof(%#v, %#v) prints %#v; want %#v`, v.level, format, msg, actual, expected)
		}
	}
}

func TestLogger_Infoln(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:INFO\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Infoln(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Infoln(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Warn(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:WARN\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Warn(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Warn(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Warnf(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	expected := "level:WARN\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Warnf(format, msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Warnf(%#v, %#v) prints %#v; want %#v`, v.level, format, msg, actual, expected)
		}
	}
}

func TestLogger_Warnln(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:WARN\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Warnln(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Warnln(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Error(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:ERROR\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Error(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Error(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Errorf(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	expected := "level:ERROR\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Errorf(format, msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Errorf(%#v, %#v) prints %#v; want %#v`, v.level, format, msg, actual, expected)
		}
	}
}

func TestLogger_Errorln(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:ERROR\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Errorln(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Errorln(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Fatal(t *testing.T) {
	t.Skip("cannot test because Logger.Fatal() calls os.Exit(1)")
}

func TestLogger_Fatalf(t *testing.T) {
	t.Skip("cannot test because Logger.Fatalf() calls os.Exit(1)")
}

func TestLogger_Fatalln(t *testing.T) {
	t.Skip("cannot test because Logger.Fatalln() calls os.Exit(1)")
}

func TestLogger_Panic(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:PANIC\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		func() {
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("log level => %v; logger.Panic(%#v) isn't calling panic()", v.level, msg)
				}
			}()
			logger.Panic(msg)
		}()
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Panic(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Panicf(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	expected := "level:PANIC\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		func() {
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("log level => %v; logger.Panicf(%#v, %#v) isn't calling panic()", v.level, format, msg)
				}
			}()
			logger.Panicf(format, msg)
		}()
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Panicf(%#v, %#v) prints %#v; want %#v`, v.level, format, msg, actual, expected)
		}
	}
}

func TestLogger_Panicln(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:PANIC\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		func() {
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("log level => %v; logger.Panicln(%#v) isn't calling panic()", v.level, msg)
				}
			}()
			logger.Panicln(msg)
		}()
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Panicln(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Print(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:NONE\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Print(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Print(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Printf(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	expected := "level:NONE\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Printf(format, msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Printf(%#v, %#v) prints %#v; want %#v`, v.level, format, msg, actual, expected)
		}
	}
}

func TestLogger_Println(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	expected := "level:NONE\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level)
		logger.Println(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Println(%#v) prints %#v; want %#v`, v.level, msg, actual, expected)
		}
	}
}

func TestLogger_Output(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	for _, currentLevel := range []log.Level{
		log.NONE,
		log.DEBUG,
		log.INFO,
		log.WARN,
		log.ERROR,
		log.FATAL,
		log.PANIC,
	} {
		for _, v := range []struct {
			level    log.Level
			expected string
		}{
			{log.NONE, "level:NONE\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"},
			{log.DEBUG, "level:DEBUG\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"},
			{log.INFO, "level:INFO\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"},
			{log.WARN, "level:WARN\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"},
			{log.ERROR, "level:ERROR\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"},
			{log.FATAL, "level:FATAL\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"},
			{log.PANIC, "level:PANIC\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\n"},
		} {
			buf.Reset()
			logger := log.New(&buf, &log.LTSVFormatter{}, currentLevel)
			logger.Output(v.level, msg)
			actual := buf.String()
			expected := v.expected
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf(`log level => %v; logger.Output(%v, %#v) prints %#v; want %#v`, currentLevel, v.level, msg, actual, expected)
			}
		}
	}
}

func TestLogger_With_Debug(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:DEBUG\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, ""},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Debug(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Debug(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Debugf(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:DEBUG\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, ""},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Debugf(format, msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Debugf(%#v, %#v) prints %#v; want %#v`, v.level, fields, format, msg, actual, expected)
		}
	}
}

func TestLogger_With_Debugln(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:DEBUG\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, ""},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Debugln(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Debugln(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Info(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:INFO\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Info(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Info(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Infof(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:INFO\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Infof(format, msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Infof(%#v, %#v) prints %#v; want %#v`, v.level, fields, format, msg, actual, expected)
		}
	}
}

func TestLogger_With_Infoln(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:INFO\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, ""},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Infoln(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Infoln(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Warn(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:WARN\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Warn(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Warn(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Warnf(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:WARN\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Warnf(format, msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Warnf(%#v, %#v) prints %#v; want %#v`, v.level, fields, format, msg, actual, expected)
		}
	}
}

func TestLogger_With_Warnln(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:WARN\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, ""},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Warnln(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Warnln(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Error(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:ERROR\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Error(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Error(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Errorf(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:ERROR\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Errorf(format, msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Errorf(%#v, %#v) prints %#v; want %#v`, v.level, fields, format, msg, actual, expected)
		}
	}
}

func TestLogger_With_Errorln(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:ERROR\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, ""},
		{log.PANIC, ""},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Errorln(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Errorln(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Fatal(t *testing.T) {
	t.Skip("cannot test because Logger.With().Fatal() calls os.Exit(1)")
}

func TestLogger_With_Fatalf(t *testing.T) {
	t.Skip("cannot test because Logger.With().Fatalf() calls os.Exit(1)")
}

func TestLogger_With_Fatalln(t *testing.T) {
	t.Skip("cannot test because Logger.With().Fatalln() calls os.Exit(1)")
}

func TestLogger_With_Panic(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:PANIC\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		func() {
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("log level => %v; logger.With(%#v).Panic(%#v) isn't calling panic()", v.level, fields, msg)
				}
			}()
			logger.Panic(msg)
		}()
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Panic(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Panicf(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:PANIC\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		func() {
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("log level => %v; logger.With(%#v).Panicf(%#v, %#v) isn't calling panic()", v.level, fields, format, msg)
				}
			}()
			logger.Panicf(format, msg)
		}()
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Panicf(%#v, %#v) prints %#v; want %#v`, v.level, fields, format, msg, actual, expected)
		}
	}
}

func TestLogger_With_Panicln(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:PANIC\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		func() {
			defer func() {
				if err := recover(); err == nil {
					t.Errorf("log level => %v; logger.With(%#v).Panicln(%#v) isn't calling panic()", v.level, fields, msg)
				}
			}()
			logger.Panicln(msg)
		}()
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Panicln(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Print(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:NONE\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Print(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Print(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Printf(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	format := "test log: %v"
	msg := "this is test"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:NONE\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log: this is test\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Printf(format, msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Printf(%#v, %#v) prints %#v; want %#v`, v.level, fields, format, msg, actual, expected)
		}
	}
}

func TestLogger_With_Println(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	expected := "level:NONE\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"
	for _, v := range []struct {
		level    log.Level
		expected string
	}{
		{log.NONE, expected},
		{log.DEBUG, expected},
		{log.INFO, expected},
		{log.WARN, expected},
		{log.ERROR, expected},
		{log.FATAL, expected},
		{log.PANIC, expected},
	} {
		buf.Reset()
		logger := log.New(&buf, &log.LTSVFormatter{}, v.level).With(fields)
		logger.Println(msg)
		actual := buf.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.With(%#v).Println(%#v) prints %#v; want %#v`, v.level, fields, msg, actual, expected)
		}
	}
}

func TestLogger_With_Output(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	msg := "test log"
	fields := log.Fields{
		"first":  1,
		"second": "two",
	}
	for _, currentLevel := range []log.Level{
		log.NONE,
		log.DEBUG,
		log.INFO,
		log.WARN,
		log.ERROR,
		log.FATAL,
		log.PANIC,
	} {
		for _, v := range []struct {
			level    log.Level
			expected string
		}{
			{log.NONE, "level:NONE\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"},
			{log.DEBUG, "level:DEBUG\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"},
			{log.INFO, "level:INFO\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"},
			{log.WARN, "level:WARN\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"},
			{log.ERROR, "level:ERROR\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"},
			{log.FATAL, "level:FATAL\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"},
			{log.PANIC, "level:PANIC\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:test log\tfirst:1\tsecond:two\n"},
		} {
			buf.Reset()
			logger := log.New(&buf, &log.LTSVFormatter{}, currentLevel).With(fields)
			logger.Output(v.level, msg)
			actual := buf.String()
			expected := v.expected
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf(`log level => %v; logger.With(%#v).Output(%v, %#v) prints %#v; want %#v`, currentLevel, fields, v.level, msg, actual, expected)
			}
		}
	}
}

func TestLogger_Level(t *testing.T) {
	for _, level := range []log.Level{
		log.NONE,
		log.DEBUG,
		log.INFO,
		log.WARN,
		log.ERROR,
		log.FATAL,
		log.PANIC,
	} {
		logger := log.New(ioutil.Discard, &log.LTSVFormatter{}, level)
		actual := logger.Level()
		expected := level
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`log level => %v; logger.Level() => %v; want %v`, level, actual, expected)
		}
	}
}

func TestLogger_SetLevel(t *testing.T) {
	now := time.Now()
	util.Now = func() time.Time { return now }
	defer func() { util.Now = time.Now }()
	var buf bytes.Buffer
	logger := log.New(&buf, &log.LTSVFormatter{}, log.DEBUG)
	var actual interface{} = logger.Level()
	var expected interface{} = log.DEBUG
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`logger.Level() => %v; want %v`, actual, expected)
	}
	msg := "test log"
	logger.Debug(msg)
	actual = buf.String()
	expected = "level:DEBUG\ttime:" + now.Format(time.RFC3339Nano) + "\tmessage:" + msg + "\n"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`logger.Debug(%#v) prints %#v; want %#v`, msg, actual, expected)
	}
	logger.SetLevel(log.INFO)
	actual = logger.Level()
	expected = log.INFO
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`After logger.SetLevel(%v); logger.Level() => %v; want %v`, log.INFO, actual, expected)
	}
	buf.Reset()
	logger.Debug(msg)
	actual = buf.String()
	expected = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`After logger.SetLevel(%v); logger.Debug(%#v) => %#v; want %#v`, log.INFO, msg, actual, expected)
	}
}

func TestLevel_String(t *testing.T) {
	for _, v := range []struct {
		level          log.Level
		name, expected string
	}{
		{log.NONE, "NONE", "NONE"},
		{log.DEBUG, "DEBUG", "DEBUG"},
		{log.INFO, "INFO", "INFO"},
		{log.WARN, "WARN", "WARN"},
		{log.ERROR, "ERROR", "ERROR"},
		{log.FATAL, "FATAL", "FATAL"},
		{log.PANIC, "PANIC", "PANIC"},
	} {
		actual := v.level.String()
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`%v.String() => %#v; want %#v`, v.name, actual, expected)
		}
	}
}

func TestFields_Get(t *testing.T) {
	fields := log.Fields{
		"first":  1,
		"second": "two",
		"third":  '3',
	}
	for _, v := range []struct {
		key      string
		expected interface{}
	}{
		{"first", 1},
		{"second", "two"},
		{"third", '3'},
		{"fourth", nil},
	} {
		var actual interface{} = fields.Get(v.key)
		var expected interface{} = v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`Fields.Get(%#v) => %#v; want %#v`, v.key, actual, expected)
		}
	}
}

func TestFields_OrderedKeys(t *testing.T) {
	fields := log.Fields{
		"second": 1,
		"first":  3,
		"ant":    4,
		"third":  2,
	}
	actual := fields.OrderedKeys()
	expected := []string{"ant", "first", "second", "third"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Fields.OrderedKeys() => %#v; want %#v`, actual, expected)
	}
}
