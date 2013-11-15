package kocha

import (
	"fmt"
	"io"
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
	var err error
	origStdout, origStderr := os.Stdout, os.Stderr
	os.Stdout, err = ioutil.TempFile("", "TestNullLogger")
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr, err = ioutil.TempFile("", "TestNullLogger")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Stdout.Close()
		os.Stderr.Close()
		os.Remove(os.Stdout.Name())
		os.Remove(os.Stderr.Name())
		os.Stdout, os.Stderr = origStdout, origStderr
	}()
	logger := NullLogger()
	logger.Output(0, "testnulllogger")
	buf, err := ioutil.ReadAll(io.MultiReader(os.Stdout, os.Stderr))
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_nullLoggerGoString(t *testing.T) {
	logger := NullLogger()
	actual := fmt.Sprintf("%#v", logger)
	expected := "kocha.NullLogger()"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestConsoleLogger(t *testing.T) {
	var err error
	origStdout, origStderr := os.Stdout, os.Stderr
	os.Stdout, err = ioutil.TempFile("", "TestConsoleLogger")
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr, err = ioutil.TempFile("", "TestConsoleLogger")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Stdout.Close()
		os.Stderr.Close()
		os.Remove(os.Stdout.Name())
		os.Remove(os.Stderr.Name())
		os.Stdout, os.Stderr = origStdout, origStderr
	}()
	logger := ConsoleLogger(-1)
	logger.Output(0, "testconsolelogger")
	buf, err := ioutil.ReadAll(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	actual := string(buf)
	expected := ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	buf, err = ioutil.ReadAll(os.Stderr)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(buf)
	expected = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	logger = ConsoleLogger(log.Ldate)
	logger.Output(0, "testconsolelogger2")
	buf, err = ioutil.ReadAll(os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(buf)
	expected = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	buf, err = ioutil.ReadAll(os.Stderr)
	if err != nil {
		t.Fatal(err)
	}
	actual = string(buf)
	expected = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_consoleLoggerGoString(t *testing.T) {
	logger := ConsoleLogger(-1)
	actual := fmt.Sprintf("%#v", logger)
	expected := fmt.Sprintf("kocha.ConsoleLogger(%d)", defaultLflag)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	logger = ConsoleLogger(log.Llongfile)
	actual = fmt.Sprintf("%#v", logger)
	expected = fmt.Sprintf("kocha.ConsoleLogger(%d)", log.Llongfile)
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
	logger.Output(0, "testlog")
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

func Test_fileLoggerGoString(t *testing.T) {
	dir, err := ioutil.TempDir("", "Test_fileLoggerGoString")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	path1 := filepath.Join(dir, "testlog1")
	logger := FileLogger(path1, -1)
	actual := fmt.Sprintf("%#v", logger)
	expected := fmt.Sprintf("kocha.FileLogger(%q, %d)", path1, defaultLflag)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	path2 := filepath.Join(dir, "testlog2")
	logger = FileLogger(path2, log.Ltime)
	actual = fmt.Sprintf("%#v", logger)
	expected = fmt.Sprintf("kocha.FileLogger(%q, %d)", path2, log.Ltime)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
