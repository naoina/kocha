package main

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

const (
	commandPrefix = "kocha-"
)

var (
	progName = filepath.Base(os.Args[0])
	aliases  = map[string]string{
		"g": "generate",
		"b": "build",
	}
)

var option struct {
	Help bool `short:"h" long:"help"`
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] COMMAND [argument...]

Commands:
    new               create a new application
    generate          generate files (alias: "g")
    build             build your application (alias: "b")
    run               run the your application
    migrate           run the migrations

Options:
    -h, --help        display this help and exit

`, progName)
}

// run runs the subcommand specified by the argument.
// run is the launcher of another command actually. It will find a subcommand
// from $GOROOT/bin, $GOPATH/bin and $PATH, and then run it.
// If subcommand is not found, run prints the usage and exit.
func run(args []string) error {
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
