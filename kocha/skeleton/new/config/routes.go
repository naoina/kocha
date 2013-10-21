package config

import (
	"../app/controllers"
	"github.com/naoina/kocha"
)

var Routes []*kocha.Route = []*kocha.Route{
	&kocha.Route{
		Name:       "root",
		Path:       "/",
		Controller: controllers.Root{},
	},
}
