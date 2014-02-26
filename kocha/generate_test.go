package main

import (
	"flag"
	"reflect"
	"testing"
)

func Test_generateCommand(t *testing.T) {
	cmd := &generateCommand{}
	var (
		actual   interface{}
		expected interface{}
	)
	actual = cmd.Name()
	expected = "generate"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = cmd.Alias()
	expected = "g"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = cmd.Short()
	expected = "generate files"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = cmd.Usage()
	expected = `generate GENERATOR [args]

Generators:

    controller
    model 
    unit  
`
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %#v, but %#v", expected, actual)
	}
	if cmd.flag != nil {
		t.Fatalf("Expect nil, but %v", cmd.flag)
	}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	cmd.DefineFlags(flags)
	flags.Parse([]string{})
	actual = cmd.flag
	expected = flags
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_generateCommand_Run_with_no_APP_PATH_given(t *testing.T) {
	cmd := &generateCommand{}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	cmd.DefineFlags(flags)
	flags.Parse([]string{})
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("Expect panic, but not occurred")
		}
	}()
	cmd.Run()
}

func Test_generateCommand_Run_with_unknown_generator(t *testing.T) {
	cmd := &generateCommand{}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	cmd.DefineFlags(flags)
	flags.Parse([]string{"unknown"})
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("Expect panic, but not occurred")
		}
	}()
	cmd.Run()
}
