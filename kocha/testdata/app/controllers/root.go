package controllers

import (
	"github.com/naoina/kocha"
)

type Root struct {
	*kocha.Controller
}

func (c *Root) GET() kocha.Result {
	return c.Render()
}
