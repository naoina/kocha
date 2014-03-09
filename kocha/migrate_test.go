package main

import (
	"flag"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func Test_migrateCommand(t *testing.T) {
	cmd := &migrateCommand{}
	for _, v := range [][]interface{}{
		{"Name", cmd.Name(), "migrate"},
		{"Alias", cmd.Alias(), ""},
		{"Short", cmd.Short(), "run the migrations"},
		{"Usage", cmd.Usage(), `migrate [-db confname] [-n n] {up|down}`},
	} {
		name, actual, expected := v[0], v[1], v[2]
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(".%v expect %v, but %v", name, expected, actual)
		}
	}

	if cmd.flag != nil {
		t.Fatalf("Expect nil, but %v", cmd.flag)
	}
	flags := flag.NewFlagSet("testflags", flag.ExitOnError)
	cmd.DefineFlags(flags)
	flags.Parse([]string{})
	actual, expected := cmd.flag, flags
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_migrateCommand_Run(t *testing.T) {
	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("(&migrateCommand{}).Run() hasn't been panicked by empty arguments")
			}
		}()
		c := &migrateCommand{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		c.DefineFlags(flags)
		flags.Parse([]string{})
		c.Run()
	}()

	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("(&migrateCommand{}).Run() hasn't been panicked by unknown direction")
			}
		}()
		c := &migrateCommand{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		c.DefineFlags(flags)
		flags.Parse([]string{"unknown"})
		c.Run()
	}()

	tester := func(expectedOutput string, args ...string) {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("(&migrateCommand{}).Run() has been panicked: %v", err)
			}
		}()
		tempDir, err := ioutil.TempDir("", "Test_migrateCommand_Run")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		appName := "testappname"
		destPath := filepath.Join(tempDir, "src", appName)
		_, filename, _, _ := runtime.Caller(0)
		baseDir := filepath.Dir(filename)
		testdataDir := filepath.Join(baseDir, "testdata")
		if err := copyAll(testdataDir, destPath); err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(destPath); err != nil {
			t.Fatal(err)
		}
		origGOPATH := build.Default.GOPATH
		defer func() {
			build.Default.GOPATH = origGOPATH
			os.Setenv("GOPATH", origGOPATH)
		}()
		build.Default.GOPATH = tempDir + string(filepath.ListSeparator) + build.Default.GOPATH
		os.Setenv("GOPATH", build.Default.GOPATH)
		c := &migrateCommand{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		c.DefineFlags(flags)
		flags.Parse(args)
		origStdout := os.Stdout
		defer func() {
			os.Stdout = origStdout
		}()
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = w
		c.Run()
		w.Close()
		out, err := ioutil.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		actual := string(out)
		expected := expectedOutput
		if !strings.HasPrefix(actual, expected) {
			t.Errorf(`(&migrateCommand{}).Run() by %q => %#v, want starts with %q`, args, actual, expected)
		}
	}

	tester("call Up: n => -1", "up")
	tester("call Down: n => -1", "down")
	tester("call Up: n => 1", "-n", "1", "up")
	tester("call Down: n => 1", "-n", "1", "down")
	tester("call Up: n => 111", "-n", "111", "up")
	tester("call Down: n => 111", "-n", "111", "down")
}
