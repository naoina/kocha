package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go/build"

	"github.com/naoina/kocha/util"
)

const (
	generatorPrefix = "kocha-generate-"
)

type generateCommand struct {
	option struct {
		Help bool `short:"h" long:"help"`
	}
}

func (c *generateCommand) Name() string {
	return "kocha generate"
}

func (c *generateCommand) Usage() string {
	return fmt.Sprintf(`Usage: %s [OPTIONS] GENERATOR [argument...]

Generate the skeleton files.

Generators:
    controller
    migration
    model
    unit

Options:
    -h, --help        display this help and exit

`, c.Name())
}

func (c *generateCommand) Option() interface{} {
	return &c.option
}

// Run execute the process for `generate` command.
func (c *generateCommand) Run(args []string) error {
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("no GENERATOR given")
	}
	name := args[0]
	var paths []string
	for _, dir := range build.Default.SrcDirs() {
		paths = append(paths, filepath.Clean(filepath.Join(filepath.Dir(dir), "bin")))
	}
	paths = append(paths, filepath.SplitList(os.Getenv("PATH"))...)
	if err := os.Setenv("PATH", strings.Join(paths, string(filepath.ListSeparator))); err != nil {
		return err
	}
	filename, err := exec.LookPath(generatorPrefix + name)
	if err != nil {
		return fmt.Errorf("could not found generator: %s", name)
	}
	cmd := exec.Command(filename, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	util.RunCommand(&generateCommand{})
}
