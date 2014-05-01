package kocha

import (
	"net/http"
	"os"
	"reflect"
	"testing"
)

func Test_newRequest(t *testing.T) {
	req, err := http.NewRequest("testMethod", "testUrl", nil)
	if err != nil {
		t.Fatal(err)
	}
	actual := newRequest(req)
	expected := &Request{
		Request: req,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestRequestScheme(t *testing.T) {
	req := &Request{}
	func() {
		os.Setenv("HTTPS", "on")
		defer os.Clearenv()
		actual := req.Scheme()
		expected := "https"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	func() {
		os.Setenv("HTTP_X_FORWARDED_SSL", "on")
		defer os.Clearenv()
		actual := req.Scheme()
		expected := "https"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	func() {
		os.Setenv("HTTP_X_FORWARDED_SCHEME", "file")
		defer os.Clearenv()
		actual := req.Scheme()
		expected := "file"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	func() {
		os.Setenv("HTTP_X_FORWARDED_PROTO", "gopher")
		defer os.Clearenv()
		actual := req.Scheme()
		expected := "gopher"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}

		os.Setenv("HTTP_X_FORWARDED_PROTO", "https, http, file")
		actual = req.Scheme()
		expected = "https"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()
}

func TestRequestIsSSL(t *testing.T) {
	req := &Request{}
	actual := req.IsSSL()
	expected := false
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	os.Setenv("HTTPS", "on")
	defer os.Clearenv()
	actual = req.IsSSL()
	expected = true
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
