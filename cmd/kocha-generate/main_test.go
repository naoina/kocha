package main

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_generateCommand_Run_withNoAPPPATHGiven(t *testing.T) {
	c := &generateCommand{}
	args := []string{}
	err := c.Run(args)
	var actual interface{} = err
	var expect interface{} = fmt.Errorf("no GENERATOR given")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
	}
}

func Test_generateCommand_Run_withUnknownGenerator(t *testing.T) {
	c := &generateCommand{}
	args := []string{"unknown"}
	err := c.Run(args)
	var actual interface{} = err
	var expect interface{} = fmt.Errorf("could not found generator: unknown")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`run(%#v) => %#v; want %#v`, args, actual, expect)
	}
}
