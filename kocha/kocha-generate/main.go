package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go/build"

	"github.com/jessevdk/go-flags"
)

const (
	progName        = "kocha generate"
	generatorPrefix = "kocha-generate-"
)

var option struct {
	Help bool `short:"h" long:"help"`
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] GENERATOR [argument...]

Generate the skeleton files.

Generators:
    controller
    migration
    model
    unit

Options:
    -h, --help        display this help and exit

`, progName)
}

// Run execute the process for `generate` command.
func run(args []string) error {
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
		fmt.Fprintf(os.Stderr, "%s: %v\n", progName, err)
		printUsage()
		os.Exit(1)
	}
}
