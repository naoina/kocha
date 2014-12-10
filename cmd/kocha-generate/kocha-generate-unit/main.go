package main

import (
	"fmt"

	"path/filepath"
	"runtime"

	"github.com/naoina/kocha/util"
)

type generateUnitCommand struct {
	option struct {
		Help bool `short:"h" long:"help"`
	}
}

func (c *generateUnitCommand) Name() string {
	return "kocha generate unit"
}

func (c *generateUnitCommand) Usage() string {
	return fmt.Sprintf(`Usage: %s [OPTIONS] NAME

Generate the skeleton files of unit.

Options:
    -h, --help        display this help and exit

`, c.Name())
}

func (c *generateUnitCommand) Option() interface{} {
	return &c.option
}

// Run generates unit skeleton files.
func (c *generateUnitCommand) Run(args []string) error {
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
		filepath.Join(skeletonDir("unit"), "unit.go"+util.TemplateSuffix),
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
	util.RunCommand(&generateUnitCommand{})
}
