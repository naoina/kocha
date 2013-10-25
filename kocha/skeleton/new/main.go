package main

import (
	"{{.appPath}}/config"
	"github.com/naoina/kocha"
)

func main() {
	kocha.Init(config.AppConfig)
	kocha.Run(config.Addr, config.Port)
}
