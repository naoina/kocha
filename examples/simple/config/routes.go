package config

import (
	"github.com/naoina/kocha"
	"github.com/naoina/kocha/examples/simple/app/controllers"
)

var Routes []*kocha.Route = []*kocha.Route{
	&kocha.Route{
		Name:       "root",
		Path:       "/",
		Controller: controllers.App{},
	},
	&kocha.Route{
		Name:       "user_show",
		Path:       "/user/:id/:name",
		Controller: controllers.User{},
	},
}
