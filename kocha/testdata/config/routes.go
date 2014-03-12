package config

import (
	"testappname/app/controllers"

	"github.com/naoina/kocha"
)

type RouteTable kocha.RouteTable

var routes = RouteTable{
	{
		Name:       "root",
		Path:       "/",
		Controller: controllers.Root{},
	},
}

func Routes() RouteTable {
	return append(routes, RouteTable{
		{
			Name:       "static",
			Path:       "/*path",
			Controller: kocha.StaticServe{},
		},
	}...)
}

func init() {
	AppConfig.Router = kocha.InitRouter(kocha.RouteTable(Routes()))
}
