package controllers

import (
	"github.com/naoina/kocha"
)

type User struct {
	kocha.Controller
}

func (c *User) Get(id int, name string) kocha.Result {
	return c.Render(kocha.Context{
		"id":   id,
		"name": name,
	})
}
