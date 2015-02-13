package kocha_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/naoina/kocha"
	"github.com/naoina/kocha/log"
	"github.com/naoina/kocha/util"
)

func TestPanicRecoverMiddleware(t *testing.T) {
	test := func(ident string, w *httptest.ResponseRecorder) {
		var actual interface{} = w.Code
		var expect interface{} = http.StatusInternalServerError
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`PanicRecoverMiddleware: %s; status => %#v; want %#v`, ident, actual, expect)
		}

		actual = w.Body.String()
		expect = "This is layout\n500 error\n\n"
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`PanicRecoverMiddleware: %s => %#v; want %#v`, ident, actual, expect)
		}

		actual = w.Header().Get("Content-Type")
		expect = "text/html"
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`PanicRecoverMiddleware: %s; Context-Type => %#v; want %#v`, ident, actual, expect)
		}
	}

	func() {
		req, err := http.NewRequest("GET", "/error", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		app.Config.Middlewares = []kocha.Middleware{
			&kocha.PanicRecoverMiddleware{},
		}
		var buf bytes.Buffer
		app.Logger = log.New(&buf, &log.LTSVFormatter{}, app.Config.Logger.Level)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		test(`GET "/error"`, w)

		actual := strings.SplitN(buf.String(), "\n", 2)[0]
		expect := "\tmessage:panic test"
		if !strings.Contains(actual, expect) {
			t.Errorf(`PanicRecoverMiddleware: GET "/error"; log => %#v; want contains => %#v`, actual, expect)
		}
	}()

	func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		app.Config.Middlewares = []kocha.Middleware{
			&kocha.PanicRecoverMiddleware{},
			&TestPanicInBeforeMiddleware{},
		}
		var buf bytes.Buffer
		app.Logger = log.New(&buf, &log.LTSVFormatter{}, app.Config.Logger.Level)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		test(`GET "/"`, w)

		actual := strings.SplitN(buf.String(), "\n", 2)[0]
		expect := "\tmessage:before"
		if !strings.Contains(actual, expect) {
			t.Errorf(`PanicRecoverMiddleware: GET "/error"; log => %#v; want contains => %#v`, actual, expect)
		}
	}()

	func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		app.Config.Middlewares = []kocha.Middleware{
			&kocha.PanicRecoverMiddleware{},
			&TestPanicInAfterMiddleware{},
		}
		var buf bytes.Buffer
		app.Logger = log.New(&buf, &log.LTSVFormatter{}, app.Config.Logger.Level)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		test(`GET "/"`, w)

		actual := strings.SplitN(buf.String(), "\n", 2)[0]
		expect := "\tmessage:after"
		if !strings.Contains(actual, expect) {
			t.Errorf(`PanicRecoverMiddleware: GET "/error"; log => %#v; want contains => %#v`, actual, expect)
		}
	}()

	func() {
		defer func() {
			actual := recover()
			expect := "before"
			if !reflect.DeepEqual(actual, expect) {
				t.Errorf(`PanicRecoverMiddleware after panic middleware: GET "/" => %#v; want %#v`, actual, expect)
			}
		}()
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		app.Config.Middlewares = []kocha.Middleware{
			&TestPanicInBeforeMiddleware{},
			&kocha.PanicRecoverMiddleware{},
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}()

	func() {
		defer func() {
			actual := recover()
			expect := "after"
			if !reflect.DeepEqual(actual, expect) {
				t.Errorf(`PanicRecoverMiddleware after panic middleware: GET "/" => %#v; want %#v`, actual, expect)
			}
		}()
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		app.Config.Middlewares = []kocha.Middleware{
			&TestPanicInAfterMiddleware{},
			&kocha.PanicRecoverMiddleware{},
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}()
}

func TestSessionMiddleware_Before(t *testing.T) {
	newRequestResponse := func(cookie *http.Cookie) (*kocha.Request, *kocha.Response) {
		r, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req := &kocha.Request{Request: r}
		if cookie != nil {
			req.AddCookie(cookie)
		}
		res := &kocha.Response{ResponseWriter: httptest.NewRecorder()}
		return req, res
	}

	origNow := util.Now
	util.Now = func() time.Time { return time.Unix(1383820443, 0) }
	defer func() {
		util.Now = origNow
	}()

	// test new session
	func() {
		app := kocha.NewTestApp()
		req, res := newRequestResponse(nil)
		c := &kocha.Context{Request: req, Response: res}
		m := &kocha.SessionMiddleware{}
		err := m.Process(app, c, func() error {
			actual := c.Session
			expected := make(kocha.Session)
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("Expect %v, but %v", expected, actual)
			}
			return fmt.Errorf("expected error")
		})
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("expected error")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`kocha.SessionMiddleware.Process(app, c, func) => %#v; want %#v`, actual, expect)
		}
		actual = c.Session
		expect = make(kocha.Session)
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf("Expect %v, but %v", expect, actual)
		}
	}()

	// test expires not found
	func() {
		app := kocha.NewTestApp()
		store := kocha.NewTestSessionCookieStore()
		sess := make(kocha.Session)
		value, err := store.Save(sess)
		if err != nil {
			t.Fatal(err)
		}
		cookie := &http.Cookie{
			Name:  app.Config.Session.Name,
			Value: value,
		}
		req, res := newRequestResponse(cookie)
		c := &kocha.Context{Request: req, Response: res}
		m := &kocha.SessionMiddleware{}
		err = m.Process(app, c, func() error {
			actual := c.Session
			expected := make(kocha.Session)
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("Expect %v, but %v", expected, actual)
			}
			return fmt.Errorf("expected error")
		})
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("expected error")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`kocha.SessionMiddleware.Process(app, c, func) => %#v; want %#v`, actual, expect)
		}
		actual = c.Session
		expect = make(kocha.Session)
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf("Expect %v, but %v", expect, actual)
		}
	}()

	// test expires invalid time format
	func() {
		app := kocha.NewTestApp()
		store := kocha.NewTestSessionCookieStore()
		sess := make(kocha.Session)
		sess[kocha.SessionExpiresKey] = "invalid format"
		value, err := store.Save(sess)
		if err != nil {
			t.Fatal(err)
		}
		cookie := &http.Cookie{
			Name:  app.Config.Session.Name,
			Value: value,
		}
		req, res := newRequestResponse(cookie)
		c := &kocha.Context{Request: req, Response: res}
		m := &kocha.SessionMiddleware{}
		err = m.Process(app, c, func() error {
			actual := c.Session
			expect := make(kocha.Session)
			if !reflect.DeepEqual(actual, expect) {
				t.Errorf("Expect %v, but %v", expect, actual)
			}
			return fmt.Errorf("expected error")
		})
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("expected error")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`kocha.SessionMiddleware.Process(app, c, func) => %#v; want %#v`, actual, expect)
		}
		actual = c.Session
		expect = make(kocha.Session)
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf("Expect %v, but %v", expect, actual)
		}
	}()

	// test expired
	func() {
		app := kocha.NewTestApp()
		store := kocha.NewTestSessionCookieStore()
		sess := make(kocha.Session)
		sess[kocha.SessionExpiresKey] = "1383820442"
		value, err := store.Save(sess)
		if err != nil {
			t.Fatal(err)
		}
		cookie := &http.Cookie{
			Name:  app.Config.Session.Name,
			Value: value,
		}
		req, res := newRequestResponse(cookie)
		c := &kocha.Context{Request: req, Response: res}
		m := &kocha.SessionMiddleware{}
		err = m.Process(app, c, func() error {
			actual := c.Session
			expected := make(kocha.Session)
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("Expect %v, but %v", expected, actual)
			}
			return fmt.Errorf("expected error")
		})
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("expected error")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`kocha.SessionMiddleware.Process(app, c, func) => %#v; want %#v`, actual, expect)
		}
		actual = c.Session
		expect = make(kocha.Session)
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf("Expect %v, but %v", expect, actual)
		}
	}()

	func() {
		app := kocha.NewTestApp()
		store := kocha.NewTestSessionCookieStore()
		sess := make(kocha.Session)
		sess[kocha.SessionExpiresKey] = "1383820443"
		sess["brown fox"] = "lazy dog"
		value, err := store.Save(sess)
		if err != nil {
			t.Fatal(err)
		}
		cookie := &http.Cookie{
			Name:  app.Config.Session.Name,
			Value: value,
		}
		req, res := newRequestResponse(cookie)
		c := &kocha.Context{Request: req, Response: res}
		m := &kocha.SessionMiddleware{}
		err = m.Process(app, c, func() error {
			return fmt.Errorf("expected error")
		})
		var actual interface{} = err
		var expect interface{} = fmt.Errorf("expected error")
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`kocha.SessionMiddlware.Process(app, c, func) => %#v; want %#v`, actual, expect)
		}
		actual = c.Session
		expect = kocha.Session{
			kocha.SessionExpiresKey: "1383820443",
			"brown fox":             "lazy dog",
		}
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf("Expect %v, but %v", expect, actual)
		}
	}()
}

func TestSessionMiddleware_After(t *testing.T) {
	app := kocha.NewTestApp()
	origNow := util.Now
	util.Now = func() time.Time { return time.Unix(1383820443, 0) }
	defer func() {
		util.Now = origNow
	}()
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req, res := &kocha.Request{Request: r}, &kocha.Response{ResponseWriter: w}
	c := &kocha.Context{Request: req, Response: res}
	c.Session = make(kocha.Session)
	app.Config.Session.SessionExpires = time.Duration(1) * time.Second
	app.Config.Session.CookieExpires = time.Duration(2) * time.Second
	m := &kocha.SessionMiddleware{}
	if err := m.Process(app, c, func() error {
		return nil
	}); err != nil {
		t.Error(err)
	}
	var (
		actual   interface{} = c.Session
		expected interface{} = kocha.Session{
			kocha.SessionExpiresKey: "1383820444", // + time.Duration(1) * time.Second
		}
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c.Session[kocha.SessionExpiresKey] = "1383820444"
	value, err := app.Config.Session.Store.Save(c.Session)
	if err != nil {
		t.Fatal(err)
	}
	c1 := res.Cookies()[0]
	c2 := &http.Cookie{
		Name:     app.Config.Session.Name,
		Value:    value,
		Path:     "/",
		Expires:  util.Now().UTC().Add(app.Config.Session.CookieExpires),
		MaxAge:   2,
		Secure:   false,
		HttpOnly: app.Config.Session.HttpOnly,
	}
	actual = c1.Name
	expected = c2.Name
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual, err = app.Config.Session.Store.Load(c1.Value)
	if err != nil {
		t.Error(err)
	}
	expected, err = app.Config.Session.Store.Load(c2.Value)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c1.Path
	expected = c2.Path
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c1.Expires
	expected = c2.Expires
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c1.MaxAge
	expected = c2.MaxAge
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c1.Secure
	expected = c2.Secure
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = c1.HttpOnly
	expected = c2.HttpOnly
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestFlashMiddleware_Before_withNilSession(t *testing.T) {
	app := kocha.NewTestApp()
	m := &kocha.FlashMiddleware{}
	c := &kocha.Context{Session: nil}
	err := m.Process(app, c, func() error {
		actual := c.Flash
		expect := kocha.Flash(nil)
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`FlashMiddleware.Process(app, c, func) => %#v; want %#v`, actual, expect)
		}
		return fmt.Errorf("expected error")
	})
	var actual interface{} = err
	var expect interface{} = fmt.Errorf("expected error")
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`kocha.FlashMiddleware.Process(app, c, func) => %#v; want %#v`, actual, expect)
	}
	actual = c.Flash
	expect = kocha.Flash(nil)
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`FlashMiddleware.Process(app, c, func) => %#v; want %#v`, actual, expect)
	}
}

func TestFlashMiddleware(t *testing.T) {
	app := kocha.NewTestApp()
	m := &kocha.FlashMiddleware{}
	c := &kocha.Context{Session: make(kocha.Session)}
	if err := m.Process(app, c, func() error {
		actual := c.Flash.Len()
		expect := 0
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`FlashMiddleware.Process(app, c, func); c.Flash.Len() => %#v; want %#v`, actual, expect)
		}
		c.Flash.Set("test_param", "abc")
		return nil
	}); err != nil {
		t.Error(err)
	}

	c.Flash = nil
	if err := m.Process(app, c, func() error {
		var actual interface{} = c.Flash.Len()
		var expected interface{} = 1
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`FlashMiddleware.Process(app, c, func) then Process(app, c, func); c.Flash.Len() => %#v; want %#v`, actual, expected)
		}
		actual = c.Flash.Get("test_param")
		expected = "abc"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`FlashMiddleware.Process(app, c, func) then Process(app, c, func); c.Flash.Get("test_param") => %#v; want %#v`, actual, expected)
		}
		return nil
	}); err != nil {
		t.Error(err)
	}

	c.Flash = nil
	if err := m.Process(app, c, func() error {
		var actual interface{} = c.Flash.Len()
		var expected interface{} = 0
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`FlashMiddleware.Process(app, c, func) then Process(app, c, func); emulated redirect; c.Flash.Len() => %#v; want %#v`, actual, expected)
		}
		actual = c.Flash.Get("test_param")
		expected = ""
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`FlashMiddleware.Process(app, c, func) then Process(app, c, func); emulated redirect; c.Flash.Get("test_param") => %#v; want %#v`, actual, expected)
		}
		return nil
	}); err != nil {
		t.Error(err)
	}
}
