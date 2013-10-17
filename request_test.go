package kocha

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNewRequest(t *testing.T) {
	req, err := http.NewRequest("testMethod", "testUrl", nil)
	if err != nil {
		t.Fatal(err)
	}
	actual := NewRequest(req)
	expected := &Request{
		Request: req,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
