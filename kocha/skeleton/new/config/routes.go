package config

import (
	"{{.appPath}}/app/controllers"
	"github.com/naoina/kocha"
)

type RouteTable kocha.RouteTable

var Routes = RouteTable{
	{
		Name:       "root",
		Path:       "/",
		Controller: controllers.Root{},
	},
}
