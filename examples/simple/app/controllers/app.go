package controllers

import (
	"github.com/naoina/kocha"
)

type App struct {
	kocha.Controller
}

func (c *App) Get() kocha.Result {
	return c.Render()
}
