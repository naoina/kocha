package main

import (
	"{{.appPath}}/config/dev"
	"github.com/naoina/kocha"
)

func main() {
	kocha.Init(dev.AppConfig)
	kocha.Run(dev.Addr, dev.Port)
}
