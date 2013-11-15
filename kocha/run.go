package main

import (
	"flag"
	"fmt"
	"github.com/naoina/kocha"
	"os"
	"os/exec"
	"path/filepath"
)

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
	c.execCmd("go", "build", "-o", appName, env+".go")
	c.execCmd(filepath.Join(dir, appName))
}

func (c *runCommand) execCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		kocha.PanicOnError(c, "abort: %v", err)
	}
}
