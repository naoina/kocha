package kocha_test

import (
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func TestFlash(t *testing.T) {
	f := kocha.Flash{}
	key := "test_key"
	var actual interface{} = f.Get(key)
	var expected interface{} = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Flash.Get(%#v) => %#v; want %#v`, key, actual, expected)
	}
	actual = f.Len()
	expected = 0
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Flash.Len() => %#v; want %#v`, actual, expected)
	}

	value := "test_value"
	f.Set(key, value)
	actual = f.Len()
	expected = 1
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Flash.Set(%#v, %#v); Flash.Len() => %#v; want %#v`, key, value, actual, expected)
	}

	actual = f.Get(key)
	expected = value
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Flash.Set(%#v, %#v); Flash.Get(%#v) => %#v; want %#v`, key, value, key, actual, expected)
	}

	key2 := "test_key2"
	value2 := "test_value2"
	f.Set(key2, value2)
	actual = f.Len()
	expected = 2
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Flash.Set(%#v, %#v); Flash.Set(%#v, %#v); Flash.Len() => %#v; want %#v`, key, value, key2, value2, actual, expected)
	}

	actual = f.Get(key2)
	expected = value2
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Flash.Set(%#v, %#v); Flash.Set(%#v, %#v); Flash.Get(%#v) => %#v; want %#v`, key, value, key2, value2, key2, actual, expected)
	}
}

func TestFlash_Get_withNil(t *testing.T) {
	f := kocha.Flash(nil)
	key := "test_key"
	var actual interface{} = f.Get(key)
	var expected interface{} = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Flash.Get(%#v) => %#v; want %#v`, key, actual, expected)
	}

	actual = f.Len()
	expected = 0
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Flash.Get(%#v); Flash.Len() => %#v; want %#v`, key, actual, expected)
	}
}

func TestFlash_Set_withNil(t *testing.T) {
	f := kocha.Flash(nil)
	key := "test_key"
	value := "test_value"
	f.Set(key, value)
	actual := f.Len()
	expected := 0
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Flash.Set(%#v, %#v); Flash.Len() => %#v; want %#v`, key, value, actual, expected)
	}
}
