package kocha_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/naoina/kocha"
)

func TestResponse_Cookies(t *testing.T) {
	w := httptest.NewRecorder()
	res := &kocha.Response{ResponseWriter: w}
	actual := res.Cookies()
	expected := []*http.Cookie(nil)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Response.Cookies() => %#v; want %#v`, actual, expected)
	}

	cookie := &http.Cookie{Name: "fox", Value: "dog"}
	res.SetCookie(cookie)
	actual = res.Cookies()
	expected = []*http.Cookie{cookie}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Response.Cookies() => %#v; want %#v`, actual, expected)
	}
}

func TestResponse_SetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	res := &kocha.Response{
		ResponseWriter: w,
	}
	cookie := &http.Cookie{
		Name:  "testCookie",
		Value: "testCookieValue",
	}
	res.SetCookie(cookie)
	actual := w.Header().Get("Set-Cookie")
	expected := cookie.String()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`Response.SetCookie(%#v) => %#v; want %#v`, cookie, actual, expected)
	}
}
