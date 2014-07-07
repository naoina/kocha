package kocha_test

import (
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func TestResourceSet_Add(t *testing.T) {
	rs := kocha.ResourceSet{}
	for _, v := range []struct {
		name string
		data interface{}
	}{
		{"text1", "test1"},
		{"text2", "test2"},
	} {
		rs.Add(v.name, v.data)
		actual := rs[v.name]
		expected := v.data
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`ResourceSet.Add("%#v", %#v) => %#v; want %#v`, v.name, v.data, actual, expected)
		}
	}
}

func TestResourceSet_Get(t *testing.T) {
	rs := kocha.ResourceSet{}
	for _, v := range []struct {
		name string
		data interface{}
	}{
		{"text1", "test1"},
		{"text2", "test2"},
	} {
		rs[v.name] = v.data
		actual := rs.Get(v.name)
		expected := v.data
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`ResourceSet.Get("%#v") => %#v; want %#v`, v.name, actual, expected)
		}
	}
}
