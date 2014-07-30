package generator

import (
	"flag"
	"path/filepath"
	"github.com/naoina/kocha/util"
)

// unitGenerator is the generator of unit.
type unitGenerator struct {
	flag *flag.FlagSet
}

// Usage returns usage of the unit generator.
func (g *unitGenerator) Usage() string {
	return "unit NAME"
}

func (g *unitGenerator) DefineFlags(fs *flag.FlagSet) {
	g.flag = fs
}

// Generate generates unit templates.
func (g *unitGenerator) Generate() {
	name := g.flag.Arg(0)
	if name == "" {
		util.PanicOnError(g, "abort: no NAME given")
	}
	camelCaseName := util.ToCamelCase(name)
	snakeCaseName := util.ToSnakeCase(name)
	data := map[string]interface{}{
		"Name": camelCaseName,
	}
	util.CopyTemplate(g,
		filepath.Join(SkeletonDir("unit"), "unit.go.template"),
		filepath.Join("app", "unit", snakeCaseName+".go"), data)
}

func init() {
	Register("unit", &unitGenerator{})
}
