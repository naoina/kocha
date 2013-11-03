package prod

import (
	"github.com/naoina/kocha"
	"{{.appPath}}/config"
)

var (
	AppName = config.AppName
	Addr = config.Addr
	Port = config.Port
	AppConfig = config.AppConfig

	logPath = "log/prod.log"
)

func init() {
	AppConfig.Logger = &kocha.Logger{
		DEBUG: kocha.Loggers{kocha.NullLogger()},
		INFO:  kocha.Loggers{kocha.NullLogger()},
		WARN:  kocha.Loggers{kocha.FileLogger(logPath, -1)},
		ERROR: kocha.Loggers{kocha.FileLogger(logPath, -1)},
	}
}
