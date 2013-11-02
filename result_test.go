package kocha

import (
	"encoding/xml"
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
	res := NewResponse(w)
	result.Proc(res)
	expected := `value1testvalue2`
	actual := w.Body.String()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestResultJSONProc(t *testing.T) {
	result := &ResultJSON{
		Context: struct{ A, B string }{"ctx1", "testctx2"},
	}
	w := httptest.NewRecorder()
	res := NewResponse(w)
	result.Proc(res)
	expected := `{"A":"ctx1","B":"testctx2"}
`
	actual := w.Body.String()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestResultXMLProc(t *testing.T) {
	result := &ResultXML{
		Context: struct {
			XMLName xml.Name `xml:"user"`
			A       string   `xml:"id"`
			B       string   `xml:"name"`
		}{A: "testId", B: "testName"},
	}
	w := httptest.NewRecorder()
	res := NewResponse(w)
	result.Proc(res)
	expected := `<user><id>testId</id><name>testName</name></user>`
	actual := w.Body.String()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestResultPlainTextProc(t *testing.T) {
	result := &ResultPlainText{"test_content"}
	w := httptest.NewRecorder()
	res := NewResponse(w)
	result.Proc(res)
	expected := `test_content`
	actual := w.Body.String()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
