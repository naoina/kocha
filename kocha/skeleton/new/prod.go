package main

import (
	"{{.appPath}}/config/prod"
	"github.com/naoina/kocha"
)

func main() {
	kocha.Init(prod.AppConfig)
	kocha.Run(prod.Addr, prod.Port)
}
