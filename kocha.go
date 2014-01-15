package kocha

import (
	"runtime"
)

const (
	DefaultHttpAddr          string = "0.0.0.0"
	DefaultHttpPort          int    = 80
	DefaultMaxClientBodySize        = 1024 * 1024 * 10 // 10MB
	StaticDir                       = "public"
)

// AppConfig represents a application-scope configuration.
type AppConfig struct {
	AppPath           string
	AppName           string
	DefaultLayout     string
	TemplateSet       TemplateSet
	Router            *Router
	Logger            *Logger
	Middlewares       []Middleware
	Session           *SessionConfig
	MaxClientBodySize int64
}

var (
	// Global logger
	Log *Logger

	// The configuration of application.
	appConfig *AppConfig

	// Whether the app has been initialized.
	initialized bool = false
)

// Init initialize the app.
func Init(config *AppConfig) {
	appConfig = config
	if appConfig.MaxClientBodySize < 1 {
		appConfig.MaxClientBodySize = DefaultMaxClientBodySize
	}
	if err := appConfig.Session.Validate(); err != nil {
		panic(err)
	}
	Log = initLogger(appConfig.Logger)
	initialized = true
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
