package main

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/naoina/kocha/util"
)

func TestGenmaiModelType_FieldTypeMap(t *testing.T) {
	m := map[string]ModelFieldType{
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
	actual := (&GenmaiModelType{}).FieldTypeMap()
	expected := m
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("(&GenmaiModelType{}).FieldTypeMap() => %#v, want %#v", actual, expected)
	}
}

func TestGenmaiModelType_TemplatePath(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(filename)
	path1, path2 := (&GenmaiModelType{}).TemplatePath()
	actual := path1
	expected := filepath.Join(basepath, "skeleton", "model", "genmai", "genmai.go"+util.TemplateSuffix)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("(&GenmaiModelType{}).TemplatePath() => %#v, $, want %#v, $", actual, expected)
	}
	actual = path2
	expected = filepath.Join(basepath, "skeleton", "model", "genmai", "config.go"+util.TemplateSuffix)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("(&GenmaiModelType{}).TemplatePath() => $, %#v, want $, %#v", actual, expected)
	}
}
