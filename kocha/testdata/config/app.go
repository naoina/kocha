package config

import (
	"path/filepath"
	"runtime"

	"github.com/naoina/kocha"
)

var (
	AppName   = "testdata"
	Addr      = "0.0.0.0"
	Port      = 9100
	AppConfig = &kocha.AppConfig{
		AppPath: rootPath,
		AppName: AppName,
		Router:  kocha.InitRouter(kocha.RouteTable(Routes)),
		TemplateSet: kocha.TemplateSetFromPaths(map[string][]string{
			AppName: []string{
				filepath.Join(rootPath, "app", "views"),
			},
		}),
	}

	_, configFileName, _, _ = runtime.Caller(0)
	rootPath                = filepath.Dir(filepath.Join(configFileName, ".."))
)

func init() {
	config := kocha.Config(AppName)
	config.Set("AppName", AppName)
	config.Set("Addr", Addr)
	config.Set("Port", Port)
}
