package kocha_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_resultContentProc(t *testing.T) {
	buf := bytes.NewBufferString("foobar")
	result := &resultContent{Body: buf}
	w := httptest.NewRecorder()
	res := newResponse(w)
	result.Proc(res)
	var actual interface{} = w.Body.String()
	var expected interface{} = "foobar"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	closer := &testCloser{bytes.NewBufferString("brown fox"), false}
	result = &resultContent{Body: closer}
	w = httptest.NewRecorder()
	res = newResponse(w)
	result.Proc(res)
	actual = w.Body.String()
	expected = "brown fox"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = closer.Closed
	expected = true
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_resultRedirectProc(t *testing.T) {
	req, err := http.NewRequest("GET", "/path/to/request", nil)
	if err != nil {
		panic(err)
	}
	result := &resultRedirect{
		Request:     newRequest(req),
		URL:         "/path/to/redirect",
		Permanently: false,
	}
	w := httptest.NewRecorder()
	res := newResponse(w)
	result.Proc(res)
	var actual interface{} = w.Code
	var expected interface{} = http.StatusFound
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = w.Header().Get("Location")
	expected = "/path/to/redirect"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	result = &resultRedirect{
		Request:     newRequest(req),
		URL:         "/path/to/redirect/permanently",
		Permanently: true,
	}
	w = httptest.NewRecorder()
	res = newResponse(w)
	result.Proc(res)
	actual = w.Code
	expected = http.StatusMovedPermanently
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = w.Header().Get("Location")
	expected = "/path/to/redirect/permanently"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

type testCloser struct {
	io.Reader
	Closed bool
}

func (c *testCloser) Close() error {
	c.Closed = true
	return nil
}
