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
	res := NewResponse(httptest.NewRecorder())
	result.Proc(res)
	expected := "text/html"
	actual := res.Header().Get("Content-Type")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	w := httptest.NewRecorder()
	res = NewResponse(w)
	result.Proc(res)
	expected = `value1testvalue2`
	actual = w.Body.String()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestResultJSONProc(t *testing.T) {
	result := &ResultJSON{
		Context: struct{ A, B string }{"ctx1", "testctx2"},
	}
	res := NewResponse(httptest.NewRecorder())
	result.Proc(res)
	expected := "application/json"
	actual := res.Header().Get("Content-Type")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	w := httptest.NewRecorder()
	res = NewResponse(w)
	res.Header().Set("Content-Type", "test/mime")
	result.Proc(res)
	expected = "test/mime"
	actual = res.Header().Get("Content-Type")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	expected = `{"A":"ctx1","B":"testctx2"}
`
	actual = w.Body.String()
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
	res := NewResponse(httptest.NewRecorder())
	result.Proc(res)
	expected := "application/xml"
	actual := res.Header().Get("Content-Type")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	w := httptest.NewRecorder()
	res = NewResponse(w)
	res.Header().Set("Content-Type", "test/mime")
	result.Proc(res)
	expected = "test/mime"
	actual = res.Header().Get("Content-Type")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	expected = `<user><id>testId</id><name>testName</name></user>`
	actual = w.Body.String()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestResultPlainTextProc(t *testing.T) {
	result := &ResultPlainText{"test_content"}
	res := NewResponse(httptest.NewRecorder())
	result.Proc(res)
	expected := "text/plain"
	actual := res.Header().Get("Content-Type")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	w := httptest.NewRecorder()
	res = NewResponse(w)
	res.Header().Set("Content-Type", "test/mime")
	result.Proc(res)
	expected = "test/mime"
	actual = res.Header().Get("Content-Type")
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	expected = `test_content`
	actual = w.Body.String()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
