package controllers

import (
	"github.com/naoina/kocha"
)

type Root struct {
	kocha.Controller
}

func (c *Root) Get() kocha.Result {
	return c.Render()
}
