package kocha

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestDefaultMiddlewares(t *testing.T) {
	actual := DefaultMiddlewares
	expected := []Middleware{
		&ResponseContentTypeMiddleware{},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestResponseContentTypeMiddlewareBefore(t *testing.T) {
	t.Skip("do nothing")
}

func TestResponseContentTypeMiddlewareAfter(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req, res := NewRequest(r), NewResponse(w)
	m := &ResponseContentTypeMiddleware{}
	actual := res.Header().Get("Content-Type")
	expected := ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	res.ContentType = "text/html"
	c := &Controller{
		Request:  req,
		Response: res,
	}
	m.After(c)
	actual = res.Header().Get("Content-Type")
	expected = "text/html; charset=utf-8"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
