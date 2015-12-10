package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/naoina/kocha/util"
)

type generateControllerCommand struct {
	option struct {
		Help bool `short:"h" long:"help"`
	}
}

func (c *generateControllerCommand) Name() string {
	return "kocha generate controller"
}

func (c *generateControllerCommand) Usage() string {
	return fmt.Sprintf(`Usage: %s [OPTIONS] NAME

Generate the skeleton files of controller.

Options:
    -h, --help        display this help and exit

`, c.Name())
}

func (c *generateControllerCommand) Option() interface{} {
	return &c.option
}

// Run generates the controller templates.
func (c *generateControllerCommand) Run(args []string) error {
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("no NAME given")
	}
	name := args[0]
	camelCaseName := util.ToCamelCase(name)
	snakeCaseName := util.ToSnakeCase(name)
	receiverName := strings.ToLower(name)
	if len(receiverName) > 1 {
		receiverName = receiverName[:2]
	} else {
		receiverName = receiverName[:1]
	}
	data := map[string]interface{}{
		"Name":     camelCaseName,
		"Receiver": receiverName,
	}
	if err := util.CopyTemplate(
		filepath.Join(skeletonDir("controller"), "controller.go"+util.TemplateSuffix),
		filepath.Join("app", "controller", snakeCaseName+".go"), data); err != nil {
		return err
	}
	if err := util.CopyTemplate(
		filepath.Join(skeletonDir("controller"), "view.html"+util.TemplateSuffix),
		filepath.Join("app", "view", snakeCaseName+".html"), data); err != nil {
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
	util.RunCommand(&generateControllerCommand{})
}
