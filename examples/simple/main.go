package main

import (
	"github.com/naoina/kocha"
	"github.com/naoina/kocha/examples/simple/config"
	"path/filepath"
	"runtime"
)

func main() {
	_, filename, _, _ := runtime.Caller(0)
	rootPath := filepath.Dir(filename)
	appName := filepath.Base(rootPath)
	kocha.Init(&kocha.AppConfig{
		AppPath: rootPath,
		AppName: appName,
		TemplateSet: kocha.TemplateSetFromPaths(map[string][]string{
			appName: []string{
				filepath.Join(rootPath, "app", "views"),
			},
		}),
		RouteTable: kocha.InitRouteTable(config.Routes),
	})
	kocha.Run("0.0.0.0", 9100)
}
