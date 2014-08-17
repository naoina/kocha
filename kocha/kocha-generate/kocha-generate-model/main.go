package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/naoina/kocha/util"
)

const (
	progName   = "kocha generate model"
	defaultORM = "genmai"
)

var modelTypeMap = map[string]ModelTyper{
	"genmai": &GenmaiModelType{},
}

var option struct {
	ORM  string `short:"o" long:"orm"`
	Help bool   `short:"h" long:"help"`
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] NAME [[field:type]...]

Options:
    -o, --orm=ORM     ORM to be used for a model [default: "%s"]
    -h, --help        display this help and exit

`, progName, defaultORM)
}

type modelField struct {
	Name       string
	Type       string
	Column     string
	OptionTags []string
}

// generate generates model templates.
func generate(args []string) error {
	if len(args) < 1 || args[0] == "" {
		return fmt.Errorf("no NAME given")
	}
	name := args[0]
	if option.ORM == "" {
		option.ORM = defaultORM
	}
	mt, exists := modelTypeMap[option.ORM]
	if !exists {
		return fmt.Errorf("unsupported ORM: `%v'", option.ORM)
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
		fmt.Fprintf(os.Stderr, "%s: %v\n", progName, err)
		printUsage()
		os.Exit(1)
	}
}
