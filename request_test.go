package kocha_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func TestRequest_Scheme(t *testing.T) {
	for _, v := range []struct {
		header string
		value  string
		expect string
	}{
		{"HTTPS", "on", "https"},
		{"X-Forwarded-SSL", "on", "https"},
		{"X-Forwarded-Scheme", "file", "file"},
		{"X-Forwarded-Proto", "gopher", "gopher"},
		{"X-Forwarded-Proto", "https, http, file", "https"},
	} {
		req := &kocha.Request{Request: &http.Request{Header: make(http.Header)}}
		req.Header.Set(v.header, v.value)
		actual := req.Scheme()
		expect := v.expect
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`Request.Scheme() with "%v: %v" => %#v; want %#v`, v.header, v.value, actual, expect)
		}
	}
}

func TestRequest_IsSSL(t *testing.T) {
	req := &kocha.Request{Request: &http.Request{Header: make(http.Header)}}
	actual := req.IsSSL()
	expected := false
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	req.Header.Set("HTTPS", "on")
	actual = req.IsSSL()
	expected = true
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestRequest_IsXHR(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req := &kocha.Request{Request: r}
	actual := req.IsXHR()
	expect := false
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`Request.IsXHR() => %#v; want %#v`, actual, expect)
	}

	req.Request.Header.Set("X-Requested-With", "XMLHttpRequest")
	actual = req.IsXHR()
	expect = true
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`Request.IsXHR() with "X-Requested-With: XMLHttpRequest" header => %#v; want %#v`, actual, expect)
	}
}
