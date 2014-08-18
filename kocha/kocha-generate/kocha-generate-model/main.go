package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/naoina/kocha/util"
)

const (
	defaultORM = "genmai"
)

var modelTypeMap = map[string]ModelTyper{
	"genmai": &GenmaiModelType{},
}

type generateModelCommand struct {
	option struct {
		ORM  string `short:"o" long:"orm"`
		Help bool   `short:"h" long:"help"`
	}
}

func (c *generateModelCommand) Name() string {
	return "kocha generate model"
}

func (c *generateModelCommand) Usage() string {
	return fmt.Sprintf(`Usage: %s [OPTIONS] NAME [[field:type]...]

Options:
    -o, --orm=ORM     ORM to be used for a model [default: "%s"]
    -h, --help        display this help and exit

`, c.Name(), defaultORM)
}

func (c *generateModelCommand) Option() interface{} {
	return &c.option
}

// Run generates model templates.
func (c *generateModelCommand) Run(args []string) error {
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("no NAME given")
	}
	name := args[0]
	if c.option.ORM == "" {
		c.option.ORM = defaultORM
	}
	mt, exists := modelTypeMap[c.option.ORM]
	if !exists {
		return fmt.Errorf("unsupported ORM: `%v'", c.option.ORM)
	}
	m := mt.FieldTypeMap()
	var fields []modelField
	for _, arg := range args[1:] {
		input := strings.Split(arg, ":")
		if len(input) != 2 {
			return fmt.Errorf("invalid argument format is specified: `%v'", arg)
		}
		name, t := input[0], input[1]
		if name == "" {
			return fmt.Errorf("field name isn't specified: `%v'", arg)
		}
		if t == "" {
			return fmt.Errorf("field type isn't specified: `%v'", arg)
		}
		ft, found := m[t]
		if !found {
			return fmt.Errorf("unsupported field type: `%v'", t)
		}
		fields = append(fields, modelField{
			Name:       util.ToCamelCase(name),
			Type:       ft.Name,
			Column:     util.ToSnakeCase(name),
			OptionTags: ft.OptionTags,
		})
	}
	camelCaseName := util.ToCamelCase(name)
	snakeCaseName := util.ToSnakeCase(name)
	data := map[string]interface{}{
		"Name":   camelCaseName,
		"Fields": fields,
	}
	templatePath, configTemplatePath := mt.TemplatePath()
	if err := util.CopyTemplate(templatePath, filepath.Join("app", "model", snakeCaseName+".go"), data); err != nil {
		return err
	}
	initPath := filepath.Join("db", "config.go")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		if err := util.CopyTemplate(configTemplatePath, initPath, nil); err != nil {
			return err
		}
	}
	return nil
}

type modelField struct {
	Name       string
	Type       string
	Column     string
	OptionTags []string
}

func main() {
	util.RunCommand(&generateModelCommand{})
}
