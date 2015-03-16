package config

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/naoina/kocha"
	"github.com/naoina/kocha/log"
)

var (
	AppName   = "testappname"
	AppConfig = &kocha.Config{
		Addr:          kocha.Getenv("KOCHA_ADDR", "127.0.0.1:9100"),
		AppPath:       rootPath,
		AppName:       AppName,
		DefaultLayout: "app",
		Template: &kocha.Template{
			PathInfo: kocha.TemplatePathInfo{
				Name: AppName,
				Paths: []string{
					filepath.Join(rootPath, "app", "view"),
				},
			},
			FuncMap: kocha.TemplateFuncMap{},
		},

		// Logger settings.
		Logger: &kocha.LoggerConfig{
			Writer:    os.Stdout,
			Formatter: &log.LTSVFormatter{},
			Level:     log.INFO,
		},

		Middlewares: []kocha.Middleware{
			&kocha.RequestLoggingMiddleware{},
			&kocha.SessionMiddleware{},
		},

		// Session settings
		Session: &kocha.SessionConfig{
			Name: "testappname_session",
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
