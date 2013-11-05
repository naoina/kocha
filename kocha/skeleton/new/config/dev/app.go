package dev

import (
	"github.com/naoina/kocha"
	"{{.appPath}}/config"
)

var (
	AppName = config.AppName
	Addr = config.Addr
	Port = config.Port
	AppConfig = config.AppConfig
)

func init() {
	AppConfig.Logger = &kocha.Logger{
		DEBUG: kocha.Loggers{kocha.ConsoleLogger(-1)},
		INFO:  kocha.Loggers{kocha.ConsoleLogger(-1)},
		WARN:  kocha.Loggers{kocha.ConsoleLogger(-1)},
		ERROR: kocha.Loggers{kocha.ConsoleLogger(-1)},
	}
	AppConfig.Middlewares = append(kocha.DefaultMiddlewares, []kocha.Middleware{}...)
}
