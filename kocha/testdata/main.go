package main

import (
	"github.com/naoina/kocha"
	"hoge/config"
)

func main() {
	kocha.Init(config.AppConfig)
	kocha.Run(config.Addr)
}
