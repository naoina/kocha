package main

import (
	"github.com/naoina/kocha"
	"testappname/config/dev"
)

func main() {
	kocha.Init(dev.AppConfig)
	kocha.Run(dev.Addr, dev.Port)
}
