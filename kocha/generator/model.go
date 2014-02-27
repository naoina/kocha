package generator

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/naoina/kocha"
)

const defaultORM = "genmai"

var typeMap = map[string]map[string]fieldType{
	"genmai": {
		"int":        fieldType{"int", nil},
		"integer":    fieldType{"int", nil},
		"int8":       fieldType{"int8", nil},
		"byte":       fieldType{"int8", nil},
		"int16":      fieldType{"int16", nil},
		"smallint":   fieldType{"int16", nil},
		"int32":      fieldType{"int32", nil},
		"int64":      fieldType{"int64", nil},
		"bigint":     fieldType{"int64", nil},
		"string":     fieldType{"string", nil},
		"text":       fieldType{"string", []string{`size:"65533"`}},
		"mediumtext": fieldType{"string", []string{`size:"16777216"`}},
		"longtext":   fieldType{"string", []string{`size:"4294967295"`}},
		"bytea":      fieldType{"[]byte", nil},
		"blob":       fieldType{"[]byte", nil},
		"mediumblob": fieldType{"[]byte", []string{`size:"65533"`}},
		"longblob":   fieldType{"[]byte", []string{`size:"4294967295"`}},
		"bool":       fieldType{"bool", nil},
		"boolean":    fieldType{"bool", nil},
		"float":      fieldType{"float64", nil},
		"float64":    fieldType{"float64", nil},
		"double":     fieldType{"float64", nil},
		"real":       fieldType{"float64", nil},
		"date":       fieldType{"time.Time", nil},
		"time":       fieldType{"time.Time", nil},
		"datetime":   fieldType{"time.Time", nil},
		"timestamp":  fieldType{"time.Time", nil},
		"decimal":    fieldType{"genmai.Rat", nil},
		"numeric":    fieldType{"genmai.Rat", nil},
	},
}

type fieldType struct {
	Name       string
	OptionTags []string
}

type modelField struct {
	Name       string
	Type       string
	Column     string
	OptionTags []string
}

// modelGenerator is the generator of model.
type modelGenerator struct {
	flag *flag.FlagSet
	orm  string
}

// Usage returns the usage of the model generator.
func (g *modelGenerator) Usage() string {
	return "model [-o ORM] NAME [[field:type] [field:type]...]"
}

func (g *modelGenerator) DefineFlags(fs *flag.FlagSet) {
	fs.StringVar(&g.orm, "o", defaultORM, fmt.Sprintf("specify ORM (default: %v)", defaultORM))
	g.flag = fs
}

// Generate generates model templates.
func (g *modelGenerator) Generate() {
	name := g.flag.Arg(0)
	if name == "" {
		kocha.PanicOnError(g, "abort: no NAME given")
	}
	m := typeMap[g.orm]
	if m == nil {
		kocha.PanicOnError(g, "abort: unsupported ORM type: `%v`", g.orm)
	}
	var fields []modelField
	for _, arg := range g.flag.Args()[1:] {
		input := strings.Split(arg, ":")
		if len(input) != 2 {
			kocha.PanicOnError(g, "abort: invalid argument format has been specified: `%v`", strings.Join(input, ", "))
		}
		name, t := input[0], input[1]
		if name == "" {
			kocha.PanicOnError(g, "abort: field name hasn't been specified")
		}
		ft, found := m[t]
		if !found {
			kocha.PanicOnError(g, "abort: unsupported field type: `%v`", t)
		}
		fields = append(fields, modelField{
			Name:       kocha.ToCamelCase(name),
			Type:       ft.Name,
			Column:     kocha.ToSnakeCase(name),
			OptionTags: ft.OptionTags,
		})
	}
	camelCaseName := kocha.ToCamelCase(name)
	snakeCaseName := kocha.ToSnakeCase(name)
	data := map[string]interface{}{
		"Name":   camelCaseName,
		"Fields": fields,
	}
	kocha.CopyTemplate(g,
		filepath.Join(SkeletonDir("model"), g.orm, g.orm+".go.template"),
		filepath.Join("app", "models", snakeCaseName+".go"), data)
	initPath := filepath.Join("db", "config.go")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		kocha.CopyTemplate(g,
			filepath.Join(SkeletonDir("model"), g.orm, "config.go.template"),
			initPath, nil)
	}
}

func init() {
	Register("model", &modelGenerator{})
}
