package main

import (
	"flag"
	"fmt"
	"github.com/howeyc/fsnotify"
	"github.com/naoina/kocha"
	"os"
	"os/exec"
	"path/filepath"
)

// Default environment
const DEFAULT_ENV = "dev"

type runCommand struct {
	flag *flag.FlagSet
}

func (c *runCommand) Name() string {
	return "run"
}

func (c *runCommand) Alias() string {
	return ""
}

func (c *runCommand) Short() string {
	return "run the your application"
}

func (c *runCommand) Usage() string {
	return fmt.Sprintf("%s ENV", c.Name())
}

func (c *runCommand) DefineFlags(fs *flag.FlagSet) {
	c.flag = fs
}

func (c *runCommand) Run() {
	env := c.flag.Arg(0)
	if env == "" {
		env = DEFAULT_ENV
		fmt.Printf("ENV is not given, use `%v`.\n", DEFAULT_ENV)
	}
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	appName := filepath.Base(dir)
	src := env + ".go"
	for {
		c.watchApp(dir, appName, src)
	}
}

func (c *runCommand) watchApp(dir, appName, src string) {
	cmd := c.execCmd("go", "build", "-o", appName, src)
	if err := cmd.Wait(); err == nil {
		cmd = c.execCmd(filepath.Join(dir, appName))
	}
	defer cmd.Process.Kill()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name()[0] == '.' {
			return filepath.SkipDir
		}
		if err := watcher.Watch(path); err != nil {
			return err
		}
		return nil
	}); err != nil {
		panic(err)
	}
	select {
	case <-watcher.Event:
	case err := <-watcher.Error:
		panic(err)
	}
	fmt.Println("Reloading...\n")
}

func (c *runCommand) execCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		kocha.PanicOnError(c, "abort: %v", err)
	}
	return cmd
}
