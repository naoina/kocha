package main

import (
	"testappname/config"

	"github.com/naoina/kocha"
)

func main() {
	kocha.Run(config.AppConfig)
}
