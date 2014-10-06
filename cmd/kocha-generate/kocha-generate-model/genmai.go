package main

import (
	"path/filepath"
	"runtime"
)

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
	templatePath = filepath.Join(skeletonDir("model"), "genmai", "genmai.go.template")
	configTemplatePath = filepath.Join(skeletonDir("model"), "genmai", "config.go.template")
	return templatePath, configTemplatePath
}

func skeletonDir(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	baseDir := filepath.Dir(filename)
	return filepath.Join(baseDir, "skeleton", name)
}
