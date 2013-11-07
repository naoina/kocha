package kocha

import (
	"fmt"
	"reflect"
	"runtime"
)

const (
	DefaultHttpAddr          string = "0.0.0.0"
	DefaultHttpPort          int    = 80
	DefaultMaxClientBodySize        = 1024 * 1024 * 10 // 10MB
)

type AppConfig struct {
	AppPath     string
	AppName     string
	TemplateSet TemplateSet
	RouteTable  RouteTable
	Logger      *Logger
	Middlewares []Middleware
	Session     SessionConfig
}

var (
	Log *Logger

	appConfig         *AppConfig
	initialized       bool  = false
	maxClientBodySize int64 = DefaultMaxClientBodySize
)

func Init(config *AppConfig) {
	appConfig = config
	if size, ok := Config(appConfig.AppName).Get("MaxClientBodySize"); ok {
		switch v := reflect.ValueOf(size); v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			maxClientBodySize = v.Int()
		default:
			panic(fmt.Errorf("`MaxClientBodySize` must be integer type. but %v", v.Type()))
		}
	}
	Log = initLogger(appConfig.Logger)
	initialized = true
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
