package generator

import (
	"flag"
)

// Map of generator.
var Generators = make(map[string]Generator)

// Generator is the interface that generator.
type Generator interface {
	Usage() string
	DefineFlags(*flag.FlagSet)
	Generate()
}

// Register register to Generators.
func Register(name string, generator Generator) {
	Generators[name] = generator
}

// Get returns Generator from Generators map.
func Get(name string) Generator {
	return Generators[name]
}
