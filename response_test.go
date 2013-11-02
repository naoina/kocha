package kocha

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewResponse(t *testing.T) {
	rw := httptest.NewRecorder()
	actual := NewResponse(rw)
	expected := &Response{
		ResponseWriter: rw,
		ContentType:    "",
		StatusCode:     http.StatusOK,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
