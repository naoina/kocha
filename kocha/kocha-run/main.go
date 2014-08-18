package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/howeyc/fsnotify"
	"github.com/jessevdk/go-flags"
	"github.com/naoina/kocha/util"
)

const (
	progName = "kocha run"
)

var option struct {
	Help bool `short:"h" long:"help"`
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS]

Run the your application.

Options:
    -h, --help        display this help and exit

`, progName)
}

func run(args []string) error {
	basedir, err := os.Getwd()
	if err != nil {
		return err
	}
	execName := filepath.Base(basedir)
	if runtime.GOOS == "windows" {
		execName += ".exe"
	}
	if err := util.PrintSettingEnv(); err != nil {
		return err
	}
	for {
		if err := watchApp(basedir, execName); err != nil {
			return err
		}
	}
}

func watchApp(basedir, execName string) error {
	cmd, err := execCmd("go", "build", "-o", execName)
	if err != nil {
		return err
	}
	if err := cmd.Wait(); err == nil {
		cmd, err = execCmd(filepath.Join(basedir, execName))
		if err != nil {
			return err
		}
	}
	defer cmd.Process.Kill()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	watchFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name()[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if err := watcher.Watch(path); err != nil {
			return err
		}
		return nil
	}
	for _, path := range []string{
		"app", "config", "main.go",
	} {
		if err := filepath.Walk(filepath.Join(basedir, path), watchFunc); err != nil {
			return err
		}
	}
	select {
	case <-watcher.Event:
	case err := <-watcher.Error:
		return err
	}
	fmt.Printf("Reloading...\n\n")
	return nil
}

func execCmd(name string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func main() {
	parser := flags.NewNamedParser(progName, flags.PrintErrors|flags.PassDoubleDash)
	if _, err := parser.AddGroup("", "", &option); err != nil {
		panic(err)
	}
	args, err := parser.Parse()
	if err != nil {
		printUsage()
		os.Exit(1)
	}
	if option.Help {
		printUsage()
		os.Exit(0)
	}
	if err := run(args); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Fprintf(os.Stderr, "%s: %v\n", progName, err)
			printUsage()
		}
		os.Exit(1)
	}
}
