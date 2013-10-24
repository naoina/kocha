package generators

import (
	"flag"
)

var Generators = make(map[string]Generator)

type Generator interface {
	Usage() string
	DefineFlags(*flag.FlagSet)
	Generate()
}

func Register(name string, generator Generator) {
	Generators[name] = generator
}

func Get(name string) Generator {
	return Generators[name]
}
