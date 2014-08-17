package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestRunWithNoAPPPATHGiven(t *testing.T) {
	args := []string{}
	err := run(args)
	var actual interface{} = err
	var expect interface{} = fmt.Errorf("no GENERATOR given")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
	}
}

func TestRunWithUnknownGenerator(t *testing.T) {
	args := []string{"unknown"}
	err := run(args)
	var actual interface{} = err
	var expect interface{} = fmt.Errorf("could not found generator: unknown")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
	}
}
