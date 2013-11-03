package kocha

import (
	"log"
	"os"
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
