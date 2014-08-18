package main

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/naoina/kocha/util"
)

const (
	commandPrefix = "kocha-"
)

var (
	aliases = map[string]string{
		"g": "generate",
		"b": "build",
	}
)

type kochaCommand struct {
	option struct {
		Help bool `short:"h" long:"help"`
	}
}

func (c *kochaCommand) Name() string {
	return filepath.Base(os.Args[0])
}

func (c *kochaCommand) Usage() string {
	return fmt.Sprintf(`Usage: %s [OPTIONS] COMMAND [argument...]

Commands:
    new               create a new application
    generate          generate files (alias: "g")
    build             build your application (alias: "b")
    run               run the your application
    migrate           run the migrations

Options:
    -h, --help        display this help and exit

`, c.Name())
}

func (c *kochaCommand) Option() interface{} {
	return &c.option
}

// run runs the subcommand specified by the argument.
// run is the launcher of another command actually. It will find a subcommand
// from $GOROOT/bin, $GOPATH/bin and $PATH, and then run it.
// If subcommand is not found, run prints the usage and exit.
func (c *kochaCommand) Run(args []string) error {
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("no COMMAND given")
	}
	var paths []string
	for _, dir := range build.Default.SrcDirs() {
		paths = append(paths, filepath.Clean(filepath.Join(filepath.Dir(dir), "bin")))
	}
	paths = append(paths, filepath.SplitList(os.Getenv("PATH"))...)
	if err := os.Setenv("PATH", strings.Join(paths, string(filepath.ListSeparator))); err != nil {
		return err
	}
	name := args[0]
	if n, exists := aliases[name]; exists {
		name = n
	}
	filename, err := exec.LookPath(commandPrefix + name)
	if err != nil {
		return fmt.Errorf("command not found: %s", name)
	}
	cmd := exec.Command(filename, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	util.RunCommand(&kochaCommand{})
}
