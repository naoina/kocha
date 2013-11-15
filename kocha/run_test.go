package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_runCommand(t *testing.T) {
	cmd := &runCommand{}
	for _, v := range [][]interface{}{
		{"Name", cmd.Name(), "run"},
		{"Alias", cmd.Alias(), ""},
		{"Short", cmd.Short(), "run the your application"},
		{"Usage", cmd.Usage(), "run ENV"},
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

func Test_runCommand_Run(t *testing.T) {
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer devnull.Close()
	origStdout, origStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = devnull, devnull
	defer func() {
		os.Stdout, os.Stderr = origStdout, origStderr
	}()
	func() {
		cmd := &runCommand{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		cmd.DefineFlags(flags)
		flags.Parse([]string{})
		defer func() {
			if err := recover(); err == nil {
				t.Error("Expect panic, but not occurred")
			}
		}()
		cmd.Run()
	}()

	func() {
		tempDir, err := ioutil.TempDir("", "Test_runCommand_Run")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		if err := ioutil.WriteFile(filepath.Join(tempDir, "dev.go"), []byte(`
package main
func main() { panic("expected panic") }
`), 0644); err != nil {
			t.Fatal(err)
		}
		cmd := &runCommand{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		cmd.DefineFlags(flags)
		flags.Parse([]string{})
		defer func() {
			if err := recover(); err == nil {
				t.Error("Expect panic, but not occurred")
			}
		}()
		cmd.Run()
	}()

	func() {
		tempDir, err := ioutil.TempDir("", "Test_runCommand_Run")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tempDir)
		if err := os.Chdir(tempDir); err != nil {
			t.Fatal(err)
		}
		if err := ioutil.WriteFile(filepath.Join(tempDir, "dev.go"), []byte(`
package main
func main() {}
`), 0644); err != nil {
			t.Fatal(err)
		}
		cmd := &runCommand{}
		flags := flag.NewFlagSet("testflags", flag.ExitOnError)
		cmd.DefineFlags(flags)
		flags.Parse([]string{})
		cmd.Run()
		binName := filepath.Base(tempDir)
		if _, err := os.Stat(filepath.Join(tempDir, binName)); err != nil {
			t.Error("Expect %v is exists, but not", binName)
		}
	}()
}
