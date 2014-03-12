package main

import (
	"testappname/config"

	"github.com/naoina/kocha"
)

func main() {
	kocha.Init(config.AppConfig)
	kocha.Run(config.Addr)
}
