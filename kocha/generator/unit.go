package generator

import (
	"flag"
	"path/filepath"
	"runtime"

	"github.com/naoina/kocha"
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
		kocha.PanicOnError(g, "abort: no NAME given")
	}
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	skeletonDir := filepath.Join(baseDir, "skeleton", "unit")
	camelCaseName := kocha.ToCamelCase(name)
	snakeCaseName := kocha.ToSnakeCase(name)
	data := map[string]interface{}{
		"Name": camelCaseName,
	}
	kocha.CopyTemplate(g,
		filepath.Join(skeletonDir, "unit.go.template"),
		filepath.Join("app", "units", snakeCaseName+".go"), data)
}

func init() {
	Register("unit", &unitGenerator{})
}
