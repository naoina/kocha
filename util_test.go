package kocha

import (
	"reflect"
	"testing"
)

func Test_toSnakeCase(t *testing.T) {
	for v, expected := range map[string]string{
		"kocha":  "kocha",
		"Kocha":  "kocha",
		"kochA":  "koch_a",
		"kOcHa":  "k_oc_ha",
		"ko_cha": "ko_cha",
	} {
		actual := toSnakeCase(v)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%v: Expect %v, but %v", v, expected, actual)
		}
	}
}

func Test_normPath(t *testing.T) {
	for v, expected := range map[string]string{
		"/":           "/",
		"/path":       "/path",
		"/path/":      "/path/",
		"//path//":    "/path/",
		"/path/to":    "/path/to",
		"/path/to///": "/path/to/",
	} {
		actual := normPath(v)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("%v: Expect %v, but %v", v, expected, actual)
		}
	}
}
