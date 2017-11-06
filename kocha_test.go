package kocha_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/naoina/kocha"
	"github.com/naoina/kocha/log"
)

type testLogFormatter struct {
}

func (f *testLogFormatter) Format(w io.Writer, entry *log.Entry) error {
	return nil
}

func newConfig() *kocha.Config {
	return &kocha.Config{
		AppPath:       "testpath",
		AppName:       "testappname",
		DefaultLayout: "testapp",
		Template:      &kocha.Template{},
		RouteTable: kocha.RouteTable{
			{
				Name:       "route1",
				Path:       "route_path1",
				Controller: &kocha.FixtureRootTestCtrl{},
			},
			{
				Name:       "route2",
				Path:       "route_path2",
				Controller: &kocha.FixtureRootTestCtrl{},
			},
		},
		Logger: &kocha.LoggerConfig{},
	}
}

func TestConst(t *testing.T) {
	for _, v := range []struct {
		name             string
		actual, expected interface{}
	}{
		{"DefaultHttpAddr", kocha.DefaultHttpAddr, "127.0.0.1:9100"},
		{"DefaultMaxClientBodySize", kocha.DefaultMaxClientBodySize, 1024 * 1024 * 10},
		{"StaticDir", kocha.StaticDir, "public"},
	} {
		actual := v.actual
		expected := v.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`%#v => %#v; want %#v`, v.name, actual, expected)
		}
	}
}

func TestNew(t *testing.T) {
	func() {
		config := newConfig()
		app, err := kocha.New(config)
		if err != nil {
			t.Fatal(err)
		}
		actual := app.Config
		expected := config
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
		if config.MaxClientBodySize != kocha.DefaultMaxClientBodySize {
			t.Errorf("Expect %v, but %v", kocha.DefaultMaxClientBodySize, config.MaxClientBodySize)
		}
	}()

	func() {
		config := newConfig()
		config.MaxClientBodySize = -1
		app, err := kocha.New(config)
		if err != nil {
			t.Fatal(err)
		}
		actual := app.Config
		expected := config
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
		if config.MaxClientBodySize != kocha.DefaultMaxClientBodySize {
			t.Errorf("Expect %v, but %v", kocha.DefaultMaxClientBodySize, config.MaxClientBodySize)
		}
	}()

	func() {
		config := newConfig()
		config.MaxClientBodySize = 20131108
		app, err := kocha.New(config)
		if err != nil {
			t.Fatal(err)
		}
		actual := app.Config
		expected := config
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %v, but %v", expected, actual)
		}
		if config.MaxClientBodySize != 20131108 {
			t.Errorf("Expect %v, but %v", 20131108, config.MaxClientBodySize)
		}
	}()

	// test for event.
	func() {
		config := newConfig()
		app, err := kocha.New(config)
		if err != nil {
			t.Fatal(err)
		}
		var actual interface{} = app.Event.WorkersPerQueue
		if actual == nil {
			t.Errorf(`New(config).Event => %#v; want not nil`, actual)
		}

		config.Event = &kocha.Event{
			WorkersPerQueue: 100,
		}
		app, err = kocha.New(config)
		if err != nil {
			t.Fatal(err)
		}
		actual = app.Event
		var expect interface{} = config.Event
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`New(config).Event => %#v; want %#v`, actual, expect)
		}
	}()
}

func TestNew_buildLogger(t *testing.T) {
	func() {
		config := newConfig()
		config.Logger = nil
		app, err := kocha.New(config)
		if err != nil {
			t.Fatal(err)
		}
		actual := app.Config.Logger
		expected := &kocha.LoggerConfig{
			Writer:    os.Stdout,
			Formatter: &log.LTSVFormatter{},
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`New(...).Config.Logger => %#v; want %#v`, actual, expected)
		}
	}()

	func() {
		var buf bytes.Buffer
		formatter := &testLogFormatter{}
		level := log.PANIC
		config := newConfig()
		config.Logger.Writer = &buf
		config.Logger.Formatter = formatter
		config.Logger.Level = level
		app, err := kocha.New(config)
		if err != nil {
			t.Fatal(err)
		}
		actual := app.Config.Logger
		expected := &kocha.LoggerConfig{
			Writer:    &buf,
			Formatter: formatter,
			Level:     level,
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`New(...).Config.Logger => %#v; want %#v`, actual, expected)
		}
	}()
}

func TestApplication_ServeHTTP(t *testing.T) {
	for _, v := range []struct {
		uri         string
		status      int
		body        string
		contentType string
	}{
		{"/", http.StatusOK, "This is layout\nThis is root\n\n", "text/html"},
		{"/user/7", http.StatusOK, "This is layout\nThis is user 7\n\n", "text/html"},
		{"/2013/07/19/user/naoina", http.StatusOK, "This is layout\nThis is date naoina: 2013-07-19\n\n", "text/html"},
		{"/missing", http.StatusNotFound, "This is layout\n404 template not found\n\n", "text/html"},
		{"/json", http.StatusOK, "{\n  \"layout\": \"application\",\n  {\"tmpl5\":\"json\"}\n\n}\n", "application/json"},
		{"/teapot", http.StatusTeapot, "This is layout\nI'm a tea pot\n\n", "text/html"},
		{"/panic_in_render", http.StatusInternalServerError, "Internal Server Error\n", "text/plain; charset=utf-8"},
		{"/static/robots.txt", http.StatusOK, "# User-Agent: *\n# Disallow: /\n", "text/plain; charset=utf-8"},
		// This returns 500 Internal Server Error (not 502 BadGateway) because the file 'error/502.html' not found.
		{"/error_controller_test", http.StatusInternalServerError, "Internal Server Error\n", "text/plain; charset=utf-8"},
	} {
		func() {
			defer func() {
				if err := recover(); err != nil {
					t.Errorf(`GET %#v is panicked; want no panic; %v`, v.uri, err)
				}
			}()
			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", v.uri, nil)
			if err != nil {
				t.Fatal(err)
			}
			app := kocha.NewTestApp()
			app.ServeHTTP(w, req)

			var actual interface{} = w.Code
			var expected interface{} = v.status
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf(`GET %#v status => %#v; want %#v`, v.uri, actual, expected)
			}

			actual = w.Body.String()
			expected = v.body
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf(`GET %#v => %#v; want %#v`, v.uri, actual, expected)
			}

			actual = w.Header().Get("Content-Type")
			expected = v.contentType
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf(`GET %#v Content-Type => %#v; want %#v`, v.uri, actual, expected)
			}
		}()
	}

	// test for panic in handler
	func() {
		defer func() {
			actual := recover()
			expect := "panic test"
			if !reflect.DeepEqual(actual, expect) {
				t.Errorf(`GET /error; recover() => %#v; want %#v`, actual, expect)
			}
		}()
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/error", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		app.ServeHTTP(w, req)
	}()

	// middleware tests
	func() {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		var called []string
		m1 := &TestMiddleware{t: t, id: "A", called: &called}
		m2 := &TestMiddleware{t: t, id: "B", called: &called}
		app.Config.Middlewares = []kocha.Middleware{m1, m2, &kocha.DispatchMiddleware{}} // all default middlewares are override
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		var actual interface{} = called
		var expected interface{} = []string{"beforeA", "beforeB", "afterB", "afterA"}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`GET "/" with middlewares calls => %#v; want %#v`, actual, expected)
		}

		actual = w.Code
		expected = http.StatusOK
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`GET "/" with middlewares status => %#v; want %#v`, actual, expected)
		}

		actual = w.Body.String()
		expected = "This is layout\nThis is root\n\n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`GET "/" with middlewares => %#v; want %#v`, actual, expected)
		}

		actual = w.Header().Get("Content-Type")
		expected = "text/html"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf(`GET "/" with middlewares Context-Type => %#v; want %#v`, actual, expected)
		}
	}()

	func() {
		defer func() {
			actual := recover()
			expect := "before"
			if !reflect.DeepEqual(actual, expect) {
				t.Errorf(`GET /; recover() => %#v; want %#v`, actual, expect)
			}
		}()
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		m := &TestPanicInBeforeMiddleware{}
		app.Config.Middlewares = []kocha.Middleware{m, &kocha.DispatchMiddleware{}} // all default middlewares are override
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}()

	func() {
		defer func() {
			actual := recover()
			expect := "after"
			if !reflect.DeepEqual(actual, expect) {
				t.Errorf(`GET /; recover() => %#v; want %#v`, actual, expect)
			}
		}()
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		m := &TestPanicInAfterMiddleware{}
		app.Config.Middlewares = []kocha.Middleware{m, &kocha.DispatchMiddleware{}} // all default middlewares are override
		app.ServeHTTP(w, req)
	}()

	// test for rewrite request url by middleware.
	func() {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("GET / with TestRewriteURLPathMiddleware has been panicked => %#v; want no panic", err)
			}
		}()
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/error", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		m := &TestRewriteURLPathMiddleware{rewritePath: "/"}
		app.Config.Middlewares = []kocha.Middleware{m, &kocha.DispatchMiddleware{}}
		app.ServeHTTP(w, req)
		var actual interface{} = w.Code
		var expect interface{} = http.StatusOK
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`GET "/" with TestRewriteURLPathMiddleware status => %#v; want %#v`, actual, expect)
		}

		actual = w.Body.String()
		expect = "This is layout\nThis is root\n\n"
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`GET "/" with TestRewriteURLPathMiddleware => %#v; want %#v`, actual, expect)
		}

		actual = w.Header().Get("Content-Type")
		expect = "text/html"
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf(`GET "/" with TestRewriteURLPathMiddleware Context-Type => %#v; want %#v`, actual, expect)
		}
	}()
}

func TestApplication_ServeHTTP_withPOST(t *testing.T) {
	// plain.
	func() {
		values := url.Values{}
		values.Set("name", "naoina")
		values.Add("type", "human")
		req, err := http.NewRequest("POST", "/post_test", bytes.NewBufferString(values.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		app := kocha.NewTestApp()
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		var actual interface{} = w.Code
		var expected interface{} = http.StatusOK
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("POST /post_test status => %#v, want %#v", actual, expected)
		}

		actual = w.Body.String()
		expected = "This is layout\nmap[params:map[]]\n\n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("POST /post_test body => %#v, want %#v", actual, expected)
		}
	}()

	// with FormMiddleware.
	func() {
		values := url.Values{}
		values.Set("name", "naoina")
		values.Add("type", "human")
		req, err := http.NewRequest("POST", "/post_test", bytes.NewBufferString(values.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		app := kocha.NewTestApp()
		app.Config.Middlewares = []kocha.Middleware{&kocha.FormMiddleware{}, &kocha.DispatchMiddleware{}}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		var actual interface{} = w.Code
		var expect interface{} = http.StatusOK
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf("POST /post_test status => %#v, want %#v", actual, expect)
		}

		actual = w.Body.String()
		expect = "This is layout\nmap[params:map[name:[naoina] type:[human]]]\n\n"
		if !reflect.DeepEqual(actual, expect) {
			t.Errorf("POST /post_test body => %#v, want %#v", actual, expect)
		}
	}()
}

type TestMiddleware struct {
	t      *testing.T
	id     string
	called *[]string
}

func (m *TestMiddleware) Process(app *kocha.Application, c *kocha.Context, next func() error) error {
	*m.called = append(*m.called, "before"+m.id)
	if err := next(); err != nil {
		return err
	}
	*m.called = append(*m.called, "after"+m.id)
	return nil
}

type TestPanicInBeforeMiddleware struct{}

func (m *TestPanicInBeforeMiddleware) Process(app *kocha.Application, c *kocha.Context, next func() error) error {
	panic("before")
	if err := next(); err != nil {
		return err
	}
	return nil
}

type TestPanicInAfterMiddleware struct{}

func (m *TestPanicInAfterMiddleware) Process(app *kocha.Application, c *kocha.Context, next func() error) error {
	if err := next(); err != nil {
		return err
	}
	panic("after")
	return nil
}

type TestRewriteURLPathMiddleware struct {
	rewritePath string
}

func (m *TestRewriteURLPathMiddleware) Process(app *kocha.Application, c *kocha.Context, next func() error) error {
	c.Request.URL.Path = m.rewritePath
	return next()
}

type testUnit struct {
	name      string
	active    bool
	callCount int
}

func (u *testUnit) ActiveIf() bool {
	u.callCount++
	return u.active
}

type testUnit2 struct{}

func (u *testUnit2) ActiveIf() bool {
	return true
}

func TestApplication_Invoke(t *testing.T) {
	// test that it invokes newFunc when ActiveIf returns true.
	func() {
		app := kocha.NewTestApp()
		unit := &testUnit{"test1", true, 0}
		called := false
		app.Invoke(unit, func() {
			called = true
		}, func() {
			t.Errorf("defaultFunc has been called")
		})
		actual := called
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %#v, but %#v", expected, actual)
		}
	}()

	// test that it invokes defaultFunc when ActiveIf returns false.
	func() {
		app := kocha.NewTestApp()
		unit := &testUnit{"test2", false, 0}
		called := false
		app.Invoke(unit, func() {
			t.Errorf("newFunc has been called")
		}, func() {
			called = true
		})
		actual := called
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %#v, but %#v", expected, actual)
		}
	}()

	// test that it invokes defaultFunc when any errors occurred in newFunc.
	func() {
		app := kocha.NewTestApp()
		unit := &testUnit{"test3", true, 0}
		called := false
		app.Invoke(unit, func() {
			panic("expected error")
		}, func() {
			called = true
		})
		actual := called
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %#v, but %#v", expected, actual)
		}
	}()

	// test that it will be panic when panic occurred in defaultFunc.
	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't occurred")
			} else if err != "expected error in defaultFunc" {
				t.Errorf("panic doesn't occurred in defaultFunc: %v", err)
			}
		}()
		app := kocha.NewTestApp()
		unit := &testUnit{"test4", false, 0}
		app.Invoke(unit, func() {
			t.Errorf("newFunc has been called")
		}, func() {
			panic("expected error in defaultFunc")
		})
	}()

	// test that it panic when panic occurred in both newFunc and defaultFunc.
	func() {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("panic doesn't occurred")
			} else if err != "expected error in defaultFunc" {
				t.Errorf("panic doesn't occurred in defaultFunc: %v", err)
			}
		}()
		app := kocha.NewTestApp()
		unit := &testUnit{"test5", true, 0}
		called := false
		app.Invoke(unit, func() {
			called = true
			panic("expected error")
		}, func() {
			panic("expected error in defaultFunc")
		})
		actual := called
		expected := true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %#v, but %#v", expected, actual)
		}
	}()

	func() {
		app := kocha.NewTestApp()
		unit := &testUnit{"test6", true, 0}
		app.Invoke(unit, func() {
			panic("expected error")
		}, func() {
			// do nothing.
		})
		var actual interface{} = unit.callCount
		var expected interface{} = 1
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %#v, but %#v", expected, actual)
		}

		// again.
		app.Invoke(unit, func() {
			t.Errorf("newFunc has been called")
		}, func() {
			// do nothing.
		})
		actual = unit.callCount
		expected = 1
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %#v, but %#v", expected, actual)
		}

		// same unit type.
		unit = &testUnit{"test7", true, 0}
		called := false
		app.Invoke(unit, func() {
			t.Errorf("newFunc has been called")
		}, func() {
			called = true
		})
		actual = called
		expected = true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %#v, but %#v", expected, actual)
		}

		// different unit type.
		unit2 := &testUnit2{}
		called = false
		app.Invoke(unit2, func() {
			called = true
		}, func() {
			t.Errorf("defaultFunc has been called")
		})
		actual = called
		expected = true
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expect %#v, but %#v", expected, actual)
		}
	}()
}

func TestGetenv(t *testing.T) {
	func() {
		actual := kocha.Getenv("TEST_KOCHA_ENV", "default value")
		expected := "default value"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Getenv(%#v, %#v) => %#v, want %#v", "TEST_KOCHA_ENV", "default value", actual, expected)
		}

		actual = os.Getenv("TEST_KOCHA_ENV")
		expected = "default value"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("os.Getenv(%#v) => %#v, want %#v", "TEST_KOCHA_ENV", actual, expected)
		}
	}()

	func() {
		os.Setenv("TEST_KOCHA_ENV", "set kocha env")
		defer os.Clearenv()
		actual := kocha.Getenv("TEST_KOCHA_ENV", "default value")
		expected := "set kocha env"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Getenv(%#v, %#v) => %#v, want %#v", "TEST_KOCHA_ENV", "default value", actual, expected)
		}

		actual = os.Getenv("TEST_KOCHA_ENV")
		expected = "set kocha env"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("os.Getenv(%#v) => %#v, want %#v", "TEST_KOCHA_ENV", actual, expected)
		}
	}()
}
