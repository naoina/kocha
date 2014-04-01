package kocha

import (
	"os"

	"github.com/joho/godotenv"
)

const (
	DefaultHttpAddr          = "127.0.0.1:9100"
	DefaultMaxClientBodySize = 1024 * 1024 * 10 // 10MB
	StaticDir                = "public"
)

// AppConfig represents a application-scope configuration.
type AppConfig struct {
	AppPath           string
	AppName           string
	DefaultLayout     string
	TemplateSet       TemplateSet
	RouteTable        RouteTable
	Logger            *Logger
	Middlewares       []Middleware
	Session           *SessionConfig
	MaxClientBodySize int64

	router      *Router
	templateMap TemplateMap
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
	templateMap, err := appConfig.TemplateSet.buildTemplateMap()
	if err != nil {
		panic(err)
	}
	appConfig.templateMap = templateMap
	router, err := appConfig.RouteTable.buildRouter()
	if err != nil {
		panic(err)
	}
	appConfig.router = router
	Log = initLogger(appConfig.Logger)
	initialized = true
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
