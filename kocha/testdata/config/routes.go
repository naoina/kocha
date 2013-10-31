package config

import (
	"github.com/naoina/kocha"
	"testappname/app/controllers"
)

type RouteTable kocha.RouteTable

var Routes = RouteTable{
	{
		Name:       "root",
		Path:       "/",
		Controller: controllers.Root{},
	},
}
