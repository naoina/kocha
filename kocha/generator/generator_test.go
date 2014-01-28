package generator

import (
	"flag"
	"reflect"
	"testing"
)

func TestGenerators(t *testing.T) {
	actual := Generators
	expected := map[string]Generator{
		"controller": &controllerGenerator{},
		"unit":       &unitGenerator{},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

type forTestGenerator struct{}

func (g *forTestGenerator) Usage() string {
	// dummy
	return ""
}

func (g *forTestGenerator) DefineFlags(fs *flag.FlagSet) {
	// dummy
}

func (g *forTestGenerator) Generate() {
	// dummy
}

func TestRegister(t *testing.T) {
	if len(Generators) != 2 {
		t.Fatalf("Expect 2, but %v", len(Generators))
	}
	expected := &forTestGenerator{}
	Register("test_generator", expected)
	actual, ok := Generators["test_generator"]
	if !ok {
		t.Fatal("Expect test_generator registered, but not")
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestGet(t *testing.T) {
	Register("test_generator", &forTestGenerator{})
	actual := Get("controller")
	expected := Generators["controller"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = Get("test_generator")
	expected = Generators["test_generator"]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
