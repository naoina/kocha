package kocha

import (
	"runtime"
)

const (
	DefaultHttpAddr string = "0.0.0.0"
	DefaultHttpPort int    = 80
)

type AppConfig struct {
	AppPath     string
	AppName     string
	TemplateSet TemplateSet
	RouteTable  []*Route
}

var (
	appConfig   *AppConfig
	initialized bool = false
)

func Init(config *AppConfig) {
	appConfig = config
	initialized = true
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
