package controller

import (
	"github.com/naoina/kocha"
)

type Root struct {
	*kocha.DefaultController
}

func (ro *Root) GET(c *kocha.Context) error {
	return kocha.Render(c, nil)
}
