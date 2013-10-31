package main

import (
	"github.com/naoina/kocha"
	"testappname/config/prod"
)

func main() {
	kocha.Init(prod.AppConfig)
	kocha.Run(prod.Addr, prod.Port)
}
