package kocha_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/naoina/kocha"
	"github.com/naoina/kocha/util"
)

func TestDefaultMiddlewares(t *testing.T) {
	actual := kocha.DefaultMiddlewares
	expected := []kocha.Middleware{
		&kocha.ResponseContentTypeMiddleware{},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf(`DefaultMiddlewares => %#v; want %#v`, actual, expected)
	}
}

func TestResponseContentTypeMiddleware_Before(t *testing.T) {
	t.Skip("do nothing")
}

func TestResponseContentTypeMiddleware_After(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req, res := &kocha.Request{Request: r}, &kocha.Response{ResponseWriter: w}
	m := &kocha.ResponseContentTypeMiddleware{}
	actual := res.Header().Get("Content-Type")
	expected := ""
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	res.ContentType = "text/html"
	c := &kocha.Controller{
		Request:  req,
		Response: res,
	}
	m.After(nil, c)
	actual = res.Header().Get("Content-Type")
	expected = "text/html"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
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
		c := &kocha.Controller{Request: req, Response: res}
		m := &kocha.SessionMiddleware{}
		m.Before(app, c)
		actual := c.Session
		expected := make(kocha.Session)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test expires not found
	func() {
		app := kocha.NewTestApp()
		store := kocha.NewTestSessionCookieStore()
		sess := make(kocha.Session)
		cookie := &http.Cookie{
			Name:  app.Config.Session.Name,
			Value: store.Save(sess),
		}
		req, res := newRequestResponse(cookie)
		c := &kocha.Controller{Request: req, Response: res}
		m := &kocha.SessionMiddleware{}
		m.Before(app, c)
		actual := c.Session
		expected := make(kocha.Session)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test expires invalid time format
	func() {
		app := kocha.NewTestApp()
		store := kocha.NewTestSessionCookieStore()
		sess := make(kocha.Session)
		sess[kocha.SessionExpiresKey] = "invalid format"
		cookie := &http.Cookie{
			Name:  app.Config.Session.Name,
			Value: store.Save(sess),
		}
		req, res := newRequestResponse(cookie)
		c := &kocha.Controller{Request: req, Response: res}
		m := &kocha.SessionMiddleware{}
		m.Before(app, c)
		actual := c.Session
		expected := make(kocha.Session)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test expired
	func() {
		app := kocha.NewTestApp()
		store := kocha.NewTestSessionCookieStore()
		sess := make(kocha.Session)
		sess[kocha.SessionExpiresKey] = "1383820442"
		cookie := &http.Cookie{
			Name:  app.Config.Session.Name,
			Value: store.Save(sess),
		}
		req, res := newRequestResponse(cookie)
		c := &kocha.Controller{Request: req, Response: res}
		m := &kocha.SessionMiddleware{}
		m.Before(app, c)
		actual := c.Session
		expected := make(kocha.Session)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	func() {
		app := kocha.NewTestApp()
		store := kocha.NewTestSessionCookieStore()
		sess := make(kocha.Session)
		sess[kocha.SessionExpiresKey] = "1383820443"
		sess["brown fox"] = "lazy dog"
		cookie := &http.Cookie{
			Name:  app.Config.Session.Name,
			Value: store.Save(sess),
		}
		req, res := newRequestResponse(cookie)
		c := &kocha.Controller{Request: req, Response: res}
		m := &kocha.SessionMiddleware{}
		m.Before(app, c)
		actual := c.Session
		expected := kocha.Session{
			kocha.SessionExpiresKey: "1383820443",
			"brown fox":             "lazy dog",
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
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
	c := &kocha.Controller{Request: req, Response: res}
	c.Session = make(kocha.Session)
	app.Config.Session.SessionExpires = time.Duration(1) * time.Second
	app.Config.Session.CookieExpires = time.Duration(2) * time.Second
	m := &kocha.SessionMiddleware{}
	m.After(app, c)
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
	c1 := res.Cookies()[0]
	c2 := &http.Cookie{
		Name:     app.Config.Session.Name,
		Value:    app.Config.Session.Store.Save(c.Session),
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
	actual = app.Config.Session.Store.Load(c1.Value)
	expected = app.Config.Session.Store.Load(c2.Value)
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
