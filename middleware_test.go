package kocha

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/naoina/kocha/util"
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
	expected = "text/html"
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestSessionMiddlewareBefore(t *testing.T) {
	newRequestResponse := func(cookie *http.Cookie) (*Request, *Response) {
		r, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		req := NewRequest(r)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		res := NewResponse(httptest.NewRecorder())
		return req, res
	}

	origNow := util.Now
	util.Now = func() time.Time { return time.Unix(1383820443, 0) }
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	defer func() {
		util.Now = origNow
		appConfig = oldAppConfig
	}()

	// test new session
	func() {
		var buf bytes.Buffer
		origLoggers := Log.INFO
		Log.INFO = Loggers{newTestBufferLogger(&buf)}
		defer func() {
			Log.INFO = origLoggers
		}()
		req, res := newRequestResponse(nil)
		c := &Controller{Request: req, Response: res}
		m := &SessionMiddleware{}
		m.Before(c)
		var (
			actual   interface{} = buf.String()
			expected interface{} = "new session\n"
		)
		if actual != expected {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
		actual = c.Session
		expected = make(Session)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test expires not found
	func() {
		var buf bytes.Buffer
		origLoggers := Log.ERROR
		Log.ERROR = Loggers{newTestBufferLogger(&buf)}
		defer func() {
			Log.ERROR = origLoggers
		}()
		store := newTestSessionCookieStore()
		sess := make(Session)
		cookie := &http.Cookie{
			Name:  appConfig.Session.Name,
			Value: store.Save(sess),
		}
		req, res := newRequestResponse(cookie)
		c := &Controller{Request: req, Response: res}
		m := &SessionMiddleware{}
		m.Before(c)
		var (
			actual   interface{} = buf.String()
			expected interface{} = "expires value not found\n"
		)
		if actual != expected {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
		actual = c.Session
		expected = make(Session)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test expires invalid time format
	func() {
		var buf bytes.Buffer
		origLoggers := Log.ERROR
		Log.ERROR = Loggers{newTestBufferLogger(&buf)}
		defer func() {
			Log.ERROR = origLoggers
		}()
		store := newTestSessionCookieStore()
		sess := make(Session)
		sess[SessionExpiresKey] = "invalid format"
		cookie := &http.Cookie{
			Name:  appConfig.Session.Name,
			Value: store.Save(sess),
		}
		req, res := newRequestResponse(cookie)
		c := &Controller{Request: req, Response: res}
		m := &SessionMiddleware{}
		m.Before(c)
		if reflect.DeepEqual(buf.Len(), 0) {
			t.Errorf("Expect %v, but %v", 0, buf.Len())
		}
		actual := c.Session
		expected := make(Session)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test expired
	func() {
		var buf bytes.Buffer
		origLoggers := Log.INFO
		Log.INFO = Loggers{newTestBufferLogger(&buf)}
		defer func() {
			Log.INFO = origLoggers
		}()
		store := newTestSessionCookieStore()
		sess := make(Session)
		sess[SessionExpiresKey] = "1383820442"
		cookie := &http.Cookie{
			Name:  appConfig.Session.Name,
			Value: store.Save(sess),
		}
		req, res := newRequestResponse(cookie)
		c := &Controller{Request: req, Response: res}
		m := &SessionMiddleware{}
		m.Before(c)
		var (
			actual   interface{} = buf.String()
			expected interface{} = "session has been expired\n"
		)
		if actual != expected {
			t.Errorf("Expect %q, but %q", expected, actual)
		}
		actual = c.Session
		expected = make(Session)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
	}()

	// test
	store := newTestSessionCookieStore()
	sess := make(Session)
	sess[SessionExpiresKey] = "1383820443"
	sess["brown fox"] = "lazy dog"
	cookie := &http.Cookie{
		Name:  appConfig.Session.Name,
		Value: store.Save(sess),
	}
	req, res := newRequestResponse(cookie)
	c := &Controller{Request: req, Response: res}
	m := &SessionMiddleware{}
	m.Before(c)
	actual := c.Session
	expected := Session{
		SessionExpiresKey: "1383820443",
		"brown fox":       "lazy dog",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
}

func TestSessionMiddlewareAfter(t *testing.T) {
	oldAppConfig := appConfig
	appConfig = newTestAppConfig()
	origNow := util.Now
	util.Now = func() time.Time { return time.Unix(1383820443, 0) }
	defer func() {
		util.Now = origNow
		appConfig = oldAppConfig
	}()
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req, res := NewRequest(r), NewResponse(w)
	c := &Controller{Request: req, Response: res}
	c.Session = make(Session)
	appConfig.Session.SessionExpires = time.Duration(1) * time.Second
	appConfig.Session.CookieExpires = time.Duration(2) * time.Second
	m := &SessionMiddleware{}
	m.After(c)
	var (
		actual   interface{} = c.Session
		expected interface{} = Session{
			SessionExpiresKey: "1383820444", // + time.Duration(1) * time.Second
		}
	)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}

	c.Session[SessionExpiresKey] = "1383820444"
	c1 := res.Cookies()[0]
	c2 := &http.Cookie{
		Name:     appConfig.Session.Name,
		Value:    appConfig.Session.Store.Save(c.Session),
		Path:     "/",
		Expires:  util.Now().UTC().Add(appConfig.Session.CookieExpires),
		MaxAge:   2,
		Secure:   false,
		HttpOnly: appConfig.Session.HttpOnly,
	}
	actual = c1.Name
	expected = c2.Name
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expect %v, but %v", expected, actual)
	}
	actual = appConfig.Session.Store.Load(c1.Value)
	expected = appConfig.Session.Store.Load(c2.Value)
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

type testBufferLogger struct{ *log.Logger }

func (l *testBufferLogger) GoString() string { return "" }
func newTestBufferLogger(buf io.Writer) logger {
	return &testBufferLogger{log.New(buf, "", 0)}
}
