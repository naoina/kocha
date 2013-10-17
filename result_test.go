package kocha

import (
	"html/template"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestResultTemplateProc(t *testing.T) {
	result := &ResultTemplate{
		Template: template.Must(template.New("test_template").Parse(`{{.key1}}test{{.key2}}`)),
		Context: Context{
			"key1": "value1",
			"key2": "value2",
		},
	}
	w := httptest.NewRecorder()
	result.Proc(w)
	expected := `value1testvalue2`
	actual := w.Body.String()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
