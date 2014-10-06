package main

// ModelTyper is an interface for a model type.
type ModelTyper interface {
	// FieldTypeMap returns type map for ORM.
	FieldTypeMap() map[string]ModelFieldType

	// TemplatePath returns paths that templates of ORM for model generation.
	TemplatePath() (templatePath string, configTemplatePath string)
}

type ModelFieldType struct {
	Name       string
	OptionTags []string
}
