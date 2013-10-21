package config

import (
	"github.com/naoina/kocha"
	"path/filepath"
	"runtime"
)

var (
	AppName   = filepath.Base(rootPath)
	Addr      = "0.0.0.0"
	Port      = 9100
	AppConfig = &kocha.AppConfig{
		AppPath:    rootPath,
		AppName:    AppName,
		RouteTable: kocha.InitRouteTable(Routes),
		TemplateSet: kocha.TemplateSetFromPaths(map[string][]string{
			AppName: []string{
				filepath.Join(rootPath, "app", "views"),
			},
		}),
	}

	_, configFileName, _, _ = runtime.Caller(0)
	rootPath                = filepath.Dir(filepath.Join(configFileName, ".."))
)
