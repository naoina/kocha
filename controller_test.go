package kocha_test

import (
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func TestMimeTypeFormats(t *testing.T) {
	var actual interface{} = len(kocha.MimeTypeFormats)
	var expected interface{} = 4
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`len(TestMimeTypeFormats) => %#v; want %#v`, actual, expected)
	}
	for k, v := range map[string]string{

		"application/json": "json",
		"application/xml":  "xml",
		"text/html":        "html",
		"text/plain":       "txt",
	} {
		if _, found := kocha.MimeTypeFormats[k]; !found {
			t.Errorf(`MimeTypeFormats["%#v"] => notfound; want %v`, k, v)
		}
	}
}

func TestMimeTypeFormats_Get(t *testing.T) {
	actual := kocha.MimeTypeFormats.Get("application/json")
	expected := "json"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	actual = kocha.MimeTypeFormats.Get("text/plain")
	expected = "txt"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestMimeTypeFormats_Set(t *testing.T) {
	mimeType := "test/mime"
	if kocha.MimeTypeFormats[mimeType] != "" {
		t.Fatalf("Expect none, but %v", kocha.MimeTypeFormats[mimeType])
	}
	expected := "testmimetype"
	kocha.MimeTypeFormats.Set(mimeType, expected)
	defer delete(kocha.MimeTypeFormats, mimeType)
	actual := kocha.MimeTypeFormats[mimeType]
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestMimeTypeFormats_Del(t *testing.T) {
	if kocha.MimeTypeFormats["text/html"] == "" {
		t.Fatal("Expect exists, but not exists")
	}
	kocha.MimeTypeFormats.Del("text/html")
	defer func() {
		kocha.MimeTypeFormats["text/html"] = "html"
	}()
	actual := kocha.MimeTypeFormats["text/html"]
	expected := ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
