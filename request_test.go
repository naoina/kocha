package kocha_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func TestRequest_Scheme(t *testing.T) {
	req := &kocha.Request{}
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

func TestRequest_IsSSL(t *testing.T) {
	req := &kocha.Request{}
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
