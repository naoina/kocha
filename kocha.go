package kocha

import (
	"runtime"
)

const (
	DefaultHttpAddr          string = "0.0.0.0"
	DefaultHttpPort          int    = 80
	DefaultMaxClientBodySize        = 1024 * 1024 * 10 // 10MB
)

type AppConfig struct {
	AppPath           string
	AppName           string
	TemplateSet       TemplateSet
	RouteTable        RouteTable
	Logger            *Logger
	Middlewares       []Middleware
	Session           SessionConfig
	MaxClientBodySize int64
}

var (
	Log *Logger

	appConfig   *AppConfig
	initialized bool = false
)

func Init(config *AppConfig) {
	appConfig = config
	if appConfig.MaxClientBodySize < 1 {
		appConfig.MaxClientBodySize = DefaultMaxClientBodySize
	}
	Log = initLogger(appConfig.Logger)
	initialized = true
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
