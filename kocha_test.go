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
		{"/", http.StatusOK, "This is layout\n\nThis is root\n\n", "text/html"},
		{"/user/7", http.StatusOK, "This is layout\n\nThis is user 7\n\n", "text/html"},
		{"/2013/07/19/user/naoina", http.StatusOK, "This is layout\n\nThis is date naoina: 2013-07-19\n\n", "text/html"},
		{"/missing", http.StatusNotFound, "Not Found", "text/plain"},
		{"/error", http.StatusInternalServerError, "This is layout\n\n500 error\n\n", "text/html"},
		{"/json", http.StatusOK, "{\n  \"layout\": \"application\",\n  \n{\"tmpl5\":\"json\"}\n\n}\n", "application/json"},
		{"/teapot", http.StatusTeapot, "This is layout\n\nI'm a tea pot\n\n", "text/html"},
		{"/panic_in_render", http.StatusInternalServerError, "Internal Server Error", "text/plain"},
		{"/static/robots.txt", http.StatusOK, "# User-Agent: *\n# Disallow: /\n", "text/plain; charset=utf-8"},
		{"/error_controller_test", http.StatusBadGateway, "Bad Gateway", "text/plain"},
	} {
		func() {
			defer func() {
				if err := recover(); err != nil {
					t.Errorf(`GET %#v is panicked; want no panic`, v.uri)
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
		app.Config.Middlewares = []kocha.Middleware{m1, m2} // all default middlewares are override
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
		expected = "This is layout\n\nThis is root\n\n"
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
			if err := recover(); err != nil {
				t.Errorf("Expect don't panic, but panic")
			}
		}()
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		m := &TestPanicInBeforeMiddleware{}
		app.Config.Middlewares = []kocha.Middleware{m} // all default middlewares are override
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		status := w.Code
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
		w := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		app := kocha.NewTestApp()
		m := &TestPanicInAfterMiddleware{}
		app.Config.Middlewares = []kocha.Middleware{m} // all default middlewares are override
		app.ServeHTTP(w, req)
		status := w.Code
		if !reflect.DeepEqual(status, http.StatusInternalServerError) {
			t.Errorf("Expect %v, but %v", http.StatusInternalServerError, status)
		}
	}()
}

func TestApplication_ServeHTTP_withPOST(t *testing.T) {
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
		expected = "This is layout\n\nmap[params:map[name:[naoina] type:[human]]]\n\n"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("POST /post_test body => %#v, want %#v", actual, expected)
		}
	}()
}

type TestMiddleware struct {
	t      *testing.T
	id     string
	called *[]string
}

func (m *TestMiddleware) Before(app *kocha.Application, c *kocha.Context) {
	*m.called = append(*m.called, "before"+m.id)
}

func (m *TestMiddleware) After(app *kocha.Application, c *kocha.Context) {
	*m.called = append(*m.called, "after"+m.id)
}

type TestPanicInBeforeMiddleware struct{}

func (m *TestPanicInBeforeMiddleware) Before(app *kocha.Application, c *kocha.Context) {
	panic("before")
}
func (m *TestPanicInBeforeMiddleware) After(app *kocha.Application, c *kocha.Context) {}

type TestPanicInAfterMiddleware struct{}

func (m *TestPanicInAfterMiddleware) Before(app *kocha.Application, c *kocha.Context) {}
func (m *TestPanicInAfterMiddleware) After(app *kocha.Application, c *kocha.Context) {
	panic("after")
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
			t.Errorf("Expect %q, but %q", expected, actual)
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
			t.Errorf("Expect %q, but %q", expected, actual)
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
			t.Errorf("Expect %q, but %q", expected, actual)
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
			t.Errorf("Expect %q, but %q", expected, actual)
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
			t.Errorf("Expect %q, but %q", expected, actual)
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
			t.Errorf("Expect %q, but %q", expected, actual)
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
			t.Errorf("Expect %q, but %q", expected, actual)
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
			t.Errorf("Expect %q, but %q", expected, actual)
		}
	}()
}

func TestSettingEnv(t *testing.T) {
	func() {
		actual := kocha.SettingEnv("TEST_KOCHA_ENV", "default value")
		expected := "default value"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("SettingEnv(%q, %q) => %q, want %q", "TEST_KOCHA_ENV", "default value", actual, expected)
		}

		actual = os.Getenv("TEST_KOCHA_ENV")
		expected = "default value"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("os.Getenv(%q) => %q, want %q", "TEST_KOCHA_ENV", actual, expected)
		}
	}()

	func() {
		os.Setenv("TEST_KOCHA_ENV", "set kocha env")
		defer os.Clearenv()
		actual := kocha.SettingEnv("TEST_KOCHA_ENV", "default value")
		expected := "set kocha env"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("SettingEnv(%q, %q) => %q, want %q", "TEST_KOCHA_ENV", "default value", actual, expected)
		}

		actual = os.Getenv("TEST_KOCHA_ENV")
		expected = "set kocha env"
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("os.Getenv(%q) => %q, want %q", "TEST_KOCHA_ENV", actual, expected)
		}
	}()
}
