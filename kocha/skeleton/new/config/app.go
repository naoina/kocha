package config

import (
	"path/filepath"
	"runtime"
	"time"

	"github.com/naoina/kocha"
)

var (
	AppName   = "{{.appName}}"
	Addr      = "127.0.0.1"
	Port      = 9100
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
			Name: "{{.appName}}_session",
			Store: &kocha.SessionCookieStore{
				// AUTO-GENERATED Random keys. DO NOT EDIT.
				SecretKey:  "{{.secretKey}}",
				SigningKey: "{{.signedKey}}",
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

	_, configFileName, _, _ = runtime.Caller(1)
	rootPath                = filepath.Dir(filepath.Join(configFileName, "..", ".."))
)

func init() {
	config := kocha.Config(AppName)
	config.Set("AppName", AppName)
	config.Set("Addr", Addr)
	config.Set("Port", Port)
}
