package config

import (
	"path/filepath"
	"runtime"
	"time"

	"github.com/naoina/kocha"
)

var (
	AppName   = "hoge"
	Addr      = kocha.SettingEnv("KOCHA_ADDR", "127.0.0.1:9100")
	Env       = kocha.SettingEnv("KOCHA_ENV", "dev") // NOTE: deprecated. will be removed in future.
	AppConfig = &kocha.AppConfig{
		AppPath:       rootPath,
		AppName:       AppName,
		DefaultLayout: "app",
		TemplateSet: kocha.TemplateSetFromPaths(map[string][]string{
			AppName: []string{
				filepath.Join(rootPath, "app", "views"),
			},
		}),

		// Session settings
		Session: &kocha.SessionConfig{
			Name: "hoge_session",
			Store: &kocha.SessionCookieStore{
				// AUTO-GENERATED Random keys. DO NOT EDIT.
				SecretKey:  "\xd4(\xd5H`\n\x17\xdbD^Kvk\x1c\xf5\xf7\x99\xf7!\xf7\x88Ll\x94\x9eg\xb5\xf3n#\x81u",
				SigningKey: "H\xa8\xb2\xa9\xbc\xd5\x18\xd9c~\xf0؉\xb5|\u007f",
			},

			// Expiration of session cookie, in seconds, from now.
			// Persistent if -1, For not specify, set 0.
			CookieExpires: time.Duration(90) * time.Hour * 24,

			// Expiration of session data, in seconds, from now.
			// Perssitent if -1, For not specify, set 0.
			SessionExpires: time.Duration(90) * time.Hour * 24,
			HttpOnly:       false,
		},

		MaxClientBodySize: 1024 * 1024 * 10, // 10MB
	}

	_, configFileName, _, _ = runtime.Caller(0)
	rootPath                = filepath.Dir(filepath.Join(configFileName, ".."))
)

func init() {
	switch Env {
	case "prod":
		AppConfig.Logger = &kocha.Logger{
			DEBUG: kocha.Loggers{kocha.NullLogger()},
			INFO:  kocha.Loggers{kocha.NullLogger()},
			WARN:  kocha.Loggers{kocha.FileLogger("log/prod.log", -1)},
			ERROR: kocha.Loggers{kocha.FileLogger("log/prod.log", -1)},
		}
		AppConfig.Middlewares = append(kocha.DefaultMiddlewares, []kocha.Middleware{
			&kocha.SessionMiddleware{},
		}...)
	default:
		AppConfig.Logger = &kocha.Logger{
			DEBUG: kocha.Loggers{kocha.ConsoleLogger(-1)},
			INFO:  kocha.Loggers{kocha.ConsoleLogger(-1)},
			WARN:  kocha.Loggers{kocha.ConsoleLogger(-1)},
			ERROR: kocha.Loggers{kocha.ConsoleLogger(-1)},
		}
		AppConfig.Middlewares = append(kocha.DefaultMiddlewares, []kocha.Middleware{
			&kocha.RequestLoggingMiddleware{},
			&kocha.SessionMiddleware{},
		}...)
	}
	config := kocha.Config(AppName)
	config.Set("AppName", AppName)
	config.Set("Addr", Addr)
}
