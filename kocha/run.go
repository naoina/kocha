package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/howeyc/fsnotify"
	"github.com/naoina/kocha"
)

const DEFAULT_KOCHA_ENV = "dev"

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
	return fmt.Sprintf("%s [KOCHA_ENV]", c.Name())
}

func (c *runCommand) DefineFlags(fs *flag.FlagSet) {
	c.flag = fs
}

func (c *runCommand) Run() {
	env := c.flag.Arg(0)
	if env == "" {
		fmt.Printf("kocha: KOCHA_ENV environment variable isn't set, use \"%v\"\n", DEFAULT_KOCHA_ENV)
		env = DEFAULT_KOCHA_ENV
	}
	os.Setenv("KOCHA_ENV", env)
	basedir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	execName := filepath.Base(basedir)
	if runtime.GOOS == "windows" {
		execName += ".exe"
	}
	for {
		c.watchApp(basedir, execName)
	}
}

func (c *runCommand) watchApp(basedir, execName string) {
	cmd := c.execCmd("go", "build", "-o", execName)
	if err := cmd.Wait(); err == nil {
		cmd = c.execCmd(filepath.Join(basedir, execName))
	}
	defer cmd.Process.Kill()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
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
			panic(err)
		}
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
