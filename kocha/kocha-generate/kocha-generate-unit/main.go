package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/jessevdk/go-flags"
	"github.com/naoina/kocha/util"
)

const (
	progName = "kocha generate unit"
)

var option struct {
	Help bool `short:"h" long:"help"`
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] NAME

Generate the skeleton files of unit.

Options:
    -h, --help        display this help and exit

`, progName)
}

// generate generates unit skeleton files.
func generate(args []string) error {
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("no NAME given")
	}
	name := args[0]
	camelCaseName := util.ToCamelCase(name)
	snakeCaseName := util.ToSnakeCase(name)
	data := map[string]interface{}{
		"Name": camelCaseName,
	}
	if err := util.CopyTemplate(
		filepath.Join(skeletonDir("unit"), "unit.go.template"),
		filepath.Join("app", "unit", snakeCaseName+".go"), data); err != nil {
		return err
	}
	return nil
}

func skeletonDir(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	return filepath.Join(baseDir, "skeleton", name)
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
	if err := generate(args); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Fprintf(os.Stderr, "%s: %v\n", progName, err)
			printUsage()
		}
		os.Exit(1)
	}
}
