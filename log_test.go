package kocha

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLogConstants(t *testing.T) {
	actual := defaultLflag
	expected := log.Ldate | log.Ltime
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestNullLogger(t *testing.T) {
	actual := NullLogger()
	expected := nullLogger
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestConsoleLogger(t *testing.T) {
	actual := ConsoleLogger(-1)
	expected := log.New(os.Stdout, "", defaultLflag)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = ConsoleLogger(log.Ldate)
	expected = log.New(os.Stdout, "", log.Ldate)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestFileLogger(t *testing.T) {
	logDir, err := ioutil.TempDir("", "TestFileLogger")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(logDir)
	logPath := filepath.Join(logDir, "testlog.log")
	logger := FileLogger(logPath, 0)
	logger.Print("testlog")
	buf, err := ioutil.ReadFile(logPath)
	if err != nil {
		panic(err)
	}
	actual := string(buf)
	expected := "testlog\n"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
