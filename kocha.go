package kocha

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sync"

	"github.com/joho/godotenv"
	"github.com/naoina/kocha/log"
	"github.com/naoina/miyabi"
)

const (
	// DefaultHttpAddr is the default listen address.
	DefaultHttpAddr = "127.0.0.1:9100"

	// DefaultMaxClientBodySize is the maximum size of HTTP request body.
	// This can be overridden by setting Config.MaxClientBodySize.
	DefaultMaxClientBodySize = 1024 * 1024 * 10 // 10MB

	// StaticDir is the directory of the static files.
	StaticDir = "public"
)

var nullMiddlewareNext = func() error {
	return nil
}

// Run starts Kocha app.
// This will launch the HTTP server by using github.com/naoina/miyabi.
// If you want to use other HTTP server that compatible with net/http such as
// http.ListenAndServe, you can use New.
func Run(config *Config) error {
	app, err := New(config)
	if err != nil {
		return err
	}
	pid := os.Getpid()
	miyabi.ServerState = func(state miyabi.State) {
		switch state {
		case miyabi.StateStart:
			fmt.Printf("Listening on %s\n", app.Config.Addr)
			fmt.Printf("Server PID: %d\n", pid)
		case miyabi.StateRestart:
			app.Logger.Warn("kocha: graceful restarted")
		case miyabi.StateShutdown:
			app.Logger.Warn("kocha: graceful shutdown")
		}
	}
	server := &miyabi.Server{
		Addr:    config.Addr,
		Handler: app,
	}
	app.Event.e.Start()
	defer app.Event.e.Stop()
	return server.ListenAndServe()
}

// Application represents a Kocha app.
// This implements the http.Handler interface.
type Application struct {
	// Config is a configuration of an application.
	Config *Config

	// Router is an HTTP request router of an application.
	Router *Router

	// Template is template sets of an application.
	Template *Template

	// Logger is an application logger.
	Logger log.Logger

	// Event is an interface of the event system.
	Event *Event

	// ResourceSet is set of resource of an application.
	ResourceSet ResourceSet

	failedUnits map[string]struct{}
	mu          sync.RWMutex
}

// New returns a new Application that configured by config.
func New(config *Config) (*Application, error) {
	app := &Application{
		Config:      config,
		failedUnits: make(map[string]struct{}),
	}
	if app.Config.Addr == "" {
		config.Addr = DefaultHttpAddr
	}
	if app.Config.MaxClientBodySize < 1 {
		config.MaxClientBodySize = DefaultMaxClientBodySize
	}
	if err := app.validateMiddlewares(); err != nil {
		return nil, err
	}
	if err := app.buildResourceSet(); err != nil {
		return nil, err
	}
	if err := app.buildTemplate(); err != nil {
		return nil, err
	}
	if err := app.buildRouter(); err != nil {
		return nil, err
	}
	if err := app.buildLogger(); err != nil {
		return nil, err
	}
	if err := app.buildEvent(); err != nil {
		return nil, err
	}
	return app, nil
}

// ServeHTTP implements the http.Handler.ServeHTTP.
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &Context{
		Layout:   app.Config.DefaultLayout,
		Data:     map[interface{}]interface{}{},
		Request:  newRequest(r),
		Response: newResponse(),
		App:      app,
		Errors:   make(map[string][]*ParamError),
	}
	defer func() {
		if err := c.Response.writeTo(w); err != nil {
			app.Logger.Error(err)
		}
	}()
	if err := app.wrapMiddlewares(c)(); err != nil {
		app.Logger.Error(err)
		c.Response.reset()
		http.Error(c.Response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// Invoke invokes newFunc.
// It invokes newFunc but will behave to fallback.
// When unit.ActiveIf returns false or any errors occurred in invoking, it invoke the defaultFunc if defaultFunc isn't nil.
// Also if any errors occurred at least once, next invoking will always invoke the defaultFunc.
func (app *Application) Invoke(unit Unit, newFunc func(), defaultFunc func()) {
	name := reflect.TypeOf(unit).String()
	defer func() {
		if err := recover(); err != nil {
			if err != ErrInvokeDefault {
				app.logStackAndError(err)
				app.mu.Lock()
				app.failedUnits[name] = struct{}{}
				app.mu.Unlock()
			}
			if defaultFunc != nil {
				defaultFunc()
			}
		}
	}()
	app.mu.RLock()
	_, failed := app.failedUnits[name]
	app.mu.RUnlock()
	if failed || !unit.ActiveIf() {
		panic(ErrInvokeDefault)
	}
	newFunc()
}

func (app *Application) buildRouter() error {
	router, err := app.Config.RouteTable.buildRouter()
	if err != nil {
		return err
	}
	app.Router = router
	return nil
}

func (app *Application) buildResourceSet() error {
	app.ResourceSet = app.Config.ResourceSet
	return nil
}

func (app *Application) buildTemplate() error {
	t, err := app.Config.Template.build(app)
	if err != nil {
		return err
	}
	app.Template = t
	return nil
}

func (app *Application) buildLogger() error {
	if app.Config.Logger == nil {
		app.Config.Logger = &LoggerConfig{}
	}
	if app.Config.Logger.Writer == nil {
		app.Config.Logger.Writer = os.Stdout
	}
	if app.Config.Logger.Formatter == nil {
		app.Config.Logger.Formatter = &log.LTSVFormatter{}
	}
	app.Logger = log.New(app.Config.Logger.Writer, app.Config.Logger.Formatter, app.Config.Logger.Level)
	return nil
}

func (app *Application) buildEvent() error {
	e, err := app.Config.Event.build(app)
	if err != nil {
		return err
	}
	app.Event = e
	return nil
}

func (app *Application) validateMiddlewares() error {
	for _, m := range app.Config.Middlewares {
		if v, ok := m.(Validator); ok {
			if err := v.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (app *Application) wrapMiddlewares(c *Context) func() error {
	wrapped := nullMiddlewareNext
	for i := len(app.Config.Middlewares) - 1; i >= 0; i-- {
		f, next := app.Config.Middlewares[i].Process, wrapped
		wrapped = func() error {
			return f(app, c, next)
		}
	}
	return wrapped
}

func (app *Application) logStackAndError(err interface{}) {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	app.Logger.Errorf("%v\n%s", err, buf[:n])
}

// Config represents a application-scope configuration.
type Config struct {
	Addr              string        // listen address, DefaultHttpAddr if empty.
	AppPath           string        // root path of the application.
	AppName           string        // name of the application.
	DefaultLayout     string        // name of the default layout.
	Template          *Template     // template config.
	RouteTable        RouteTable    // routing config.
	Middlewares       []Middleware  // middlewares.
	Logger            *LoggerConfig // logger config.
	Event             *Event        // event config.
	MaxClientBodySize int64         // maximum size of request body, DefaultMaxClientBodySize if 0

	ResourceSet ResourceSet
}

// SettingEnv is similar to os.Getenv.
// However, SettingEnv returns def value if the variable is not present, and
// sets def to environment variable.
func SettingEnv(key, def string) string {
	env := os.Getenv(key)
	if env != "" {
		return env
	}
	os.Setenv(key, def)
	return def
}

func init() {
	_ = godotenv.Load()
}
