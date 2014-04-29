package kocha

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
)

func TestServer(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler(w, req)
	status := w.Code
	if !reflect.DeepEqual(status, http.StatusOK) {
		t.Errorf("Expect %v, but %v", http.StatusOK, status)
	}
	body := w.Body.String()
	expected := "tmpl1"
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Expect %v, but %v", expected, body)
	}
	actual := w.Header().Get("Content-Type")
	expected = "text/html"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/user/7", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler(w, req)
	status = w.Code
	if !reflect.DeepEqual(status, http.StatusOK) {
		t.Errorf("Expect %v, but %v", http.StatusOK, status)
	}
	body = w.Body.String()
	expected = "tmpl2-7"
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Expect %v, but %v", expected, body)
	}

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/2013/07/19/user/naoina", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler(w, req)
	status = w.Code
	if !reflect.DeepEqual(status, http.StatusOK) {
		t.Errorf("Expect %v, but %v", http.StatusOK, status)
	}
	body = w.Body.String()
	expected = "tmpl3-naoina-2013-7-19"
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Expect %v, but %v", expected, body)
	}

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/missing", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler(w, req)
	status = w.Code
	if !reflect.DeepEqual(status, http.StatusNotFound) {
		t.Errorf("Expect %v, but %v", http.StatusNotFound, status)
	}

	func() {
		log.SetOutput(ioutil.Discard)
		defer log.SetOutput(os.Stdout)
		w = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/error", nil)
		if err != nil {
			t.Fatal(err)
		}
		handler(w, req)
		status = w.Code
		if !reflect.DeepEqual(status, http.StatusInternalServerError) {
			t.Errorf("Expect %v, but %v", http.StatusInternalServerError, status)
		}
	}()

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler(w, req)
	status = w.Code
	if !reflect.DeepEqual(status, http.StatusOK) {
		t.Errorf("Expect %v, but %v", http.StatusOK, status)
	}
	body = w.Body.String()
	expected = `{"tmpl5":"json"}`
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Expect %v, but %v", expected, body)
	}
	actual = w.Header().Get("Content-Type")
	expected = "application/json"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/teapot", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler(w, req)
	status = w.Code
	if !reflect.DeepEqual(status, http.StatusTeapot) {
		t.Errorf("Expect %v, but %v", http.StatusTeapot, status)
	}
	body = w.Body.String()
	expected = `teapot`
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Expect %v, but %v", expected, body)
	}
	if !reflect.DeepEqual(w.Code, http.StatusTeapot) {
		t.Errorf("Expect %v, but %v", http.StatusTeapot, w.Code)
	}

	func() {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("Expect doesn't panic, but panic")
			}
		}()
		w = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/panic_in_render", nil)
		if err != nil {
			t.Fatal(err)
		}
		handler(w, req)
		status = w.Code
		if !reflect.DeepEqual(status, http.StatusInternalServerError) {
			t.Errorf("Expect %v, but %v", http.StatusInternalServerError)
		}
	}()

	// middleware tests
	w = httptest.NewRecorder()
	req, err = http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	m := &TestMiddleware{t: t}
	appConfig.Middlewares = []Middleware{m} // all default middlewares are override
	handler(w, req)
	if !reflect.DeepEqual(m.called, "beforeafter") {
		t.Errorf("Expect %v, but %v", "beforeafter", m.called)
	}
	status = w.Code
	if !reflect.DeepEqual(status, http.StatusOK) {
		t.Errorf("Expect %v, but %v", http.StatusOK, status)
	}
	body = w.Body.String()
	expected = "tmpl1"
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Expect %v, but %v", expected, body)
	}
	actual = w.Header().Get("Content-Type")
	expected = ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	func() {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("Expect don't panic, but panic")
			}
		}()
		w = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		m := &TestPanicInBeforeMiddleware{}
		appConfig.Middlewares = []Middleware{m} // all default middlewares are override
		handler(w, req)
		status = w.Code
		if !reflect.DeepEqual(status, http.StatusInternalServerError) {
			t.Errorf("Expect %v, but %v", http.StatusInternalServerError, status)
		}
	}()

	func() {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("Expect don't panic, but panic")
			}
		}()
		w = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		m := &TestPanicInAfterMiddleware{}
		appConfig.Middlewares = []Middleware{m} // all default middlewares are override
		handler(w, req)
		status = w.Code
		if !reflect.DeepEqual(status, http.StatusInternalServerError) {
			t.Errorf("Expect %v, but %v", http.StatusInternalServerError, status)
		}
	}()
}

func TestServer_WithPOST(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		appConfig = oldAppConfig
	}()

	func() {
		values := url.Values{}
		values.Set("name", "naoina")
		values.Add("type", "human")
		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/post_test", bytes.NewBufferString(values.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		handler(w, req)
		var actual interface{} = w.Code
		var expected interface{} = http.StatusOK
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("POST /post_test status => %#v, want %#v", actual, expected)
		}

		actual = w.Body.String()
		expected = `tmpl7-map[name:[naoina] type:[human]]`
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("POST /post_test body => %#v, want %#v", actual, expected)
		}
	}()
}

type TestMiddleware struct {
	t      *testing.T
	called string
}

func (m *TestMiddleware) Before(c *Controller) {
	m.called += "before"
}

func (m *TestMiddleware) After(c *Controller) {
	m.called += "after"
}

type TestPanicInBeforeMiddleware struct{}

func (m *TestPanicInBeforeMiddleware) Before(c *Controller) { panic("before") }
func (m *TestPanicInBeforeMiddleware) After(c *Controller)  {}

type TestPanicInAfterMiddleware struct{}

func (m *TestPanicInAfterMiddleware) Before(c *Controller) {}
func (m *TestPanicInAfterMiddleware) After(c *Controller)  { panic("after") }
