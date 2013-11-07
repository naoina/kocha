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

func Test_Response_Cookies(t *testing.T) {
	w := httptest.NewRecorder()
	res := &Response{ResponseWriter: w}
	actual := res.Cookies()
	expected := res.cookies
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	cookie := &http.Cookie{Name: "fox", Value: "dog"}
	res.cookies = append(res.cookies, cookie)
	actual = res.Cookies()
	expected = res.cookies
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func Test_Response_SetCookie(t *testing.T) {
	w := httptest.NewRecorder()
	res := &Response{
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
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}
