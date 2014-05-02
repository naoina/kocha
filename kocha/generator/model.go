package generator

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/naoina/kocha/util"
)

const DefaultORM = "genmai"

var modelTypeMap = make(map[string]ModelTyper)

// ModelTyper is an interface for a model type.
type ModelTyper interface {
	// FieldTypeMap returns type map for ORM.
	FieldTypeMap() map[string]ModelFieldType

	// TemplatePath returns paths that templates of ORM for model generation.
	TemplatePath() (templatePath string, configTemplatePath string)
}

// GenmaiModelType implements ModelTyper interface.
type GenmaiModelType struct{}

// FieldTypeMap returns type map for Genmai ORM.
func (mt *GenmaiModelType) FieldTypeMap() map[string]ModelFieldType {
	return map[string]ModelFieldType{
		"int":        ModelFieldType{"int", nil},
		"integer":    ModelFieldType{"int", nil},
		"int8":       ModelFieldType{"int8", nil},
		"byte":       ModelFieldType{"int8", nil},
		"int16":      ModelFieldType{"int16", nil},
		"smallint":   ModelFieldType{"int16", nil},
		"int32":      ModelFieldType{"int32", nil},
		"int64":      ModelFieldType{"int64", nil},
		"bigint":     ModelFieldType{"int64", nil},
		"string":     ModelFieldType{"string", nil},
		"text":       ModelFieldType{"string", []string{`size:"65533"`}},
		"mediumtext": ModelFieldType{"string", []string{`size:"16777216"`}},
		"longtext":   ModelFieldType{"string", []string{`size:"4294967295"`}},
		"bytea":      ModelFieldType{"[]byte", nil},
		"blob":       ModelFieldType{"[]byte", nil},
		"mediumblob": ModelFieldType{"[]byte", []string{`size:"65533"`}},
		"longblob":   ModelFieldType{"[]byte", []string{`size:"4294967295"`}},
		"bool":       ModelFieldType{"bool", nil},
		"boolean":    ModelFieldType{"bool", nil},
		"float":      ModelFieldType{"genmai.Float64", nil},
		"float64":    ModelFieldType{"genmai.Float64", nil},
		"double":     ModelFieldType{"genmai.Float64", nil},
		"real":       ModelFieldType{"genmai.Float64", nil},
		"date":       ModelFieldType{"time.Time", nil},
		"time":       ModelFieldType{"time.Time", nil},
		"datetime":   ModelFieldType{"time.Time", nil},
		"timestamp":  ModelFieldType{"time.Time", nil},
		"decimal":    ModelFieldType{"genmai.Rat", nil},
		"numeric":    ModelFieldType{"genmai.Rat", nil},
	}
}

// TemplatePath returns paths that templates of Genmai ORM for model generation.
func (mt *GenmaiModelType) TemplatePath() (templatePath string, configTemplatePath string) {
	templatePath = filepath.Join(SkeletonDir("model"), "genmai", "genmai.go.template")
	configTemplatePath = filepath.Join(SkeletonDir("model"), "genmai", "config.go.template")
	return templatePath, configTemplatePath
}

type ModelFieldType struct {
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
	fs.StringVar(&g.orm, "o", DefaultORM, fmt.Sprintf("specify ORM (default: %v)", DefaultORM))
	g.flag = fs
}

// Generate generates model templates.
func (g *modelGenerator) Generate() {
	name := g.flag.Arg(0)
	if name == "" {
		util.PanicOnError(g, "abort: no NAME given")
	}
	mt := modelTypeMap[g.orm]
	if mt == nil {
		util.PanicOnError(g, "abort: unsupported ORM type: `%v`", g.orm)
	}
	m := mt.FieldTypeMap()
	var fields []modelField
	for _, arg := range g.flag.Args()[1:] {
		input := strings.Split(arg, ":")
		if len(input) != 2 {
			util.PanicOnError(g, "abort: invalid argument format has been specified: `%v`", strings.Join(input, ", "))
		}
		name, t := input[0], input[1]
		if name == "" {
			util.PanicOnError(g, "abort: field name hasn't been specified")
		}
		ft, found := m[t]
		if !found {
			util.PanicOnError(g, "abort: unsupported field type: `%v`", t)
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
	util.CopyTemplate(g, templatePath, filepath.Join("app", "models", snakeCaseName+".go"), data)
	initPath := filepath.Join("db", "config.go")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		util.CopyTemplate(g, configTemplatePath, initPath, nil)
	}
}

// RegisterModelType registers an ORM-specific model type.
// If already registered, it overwrites.
func RegisterModelType(name string, mt ModelTyper) {
	modelTypeMap[name] = mt
}

func init() {
	RegisterModelType("genmai", &GenmaiModelType{})
	Register("model", &modelGenerator{})
}
