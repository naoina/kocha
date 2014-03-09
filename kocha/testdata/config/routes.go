package config

import (
	"github.com/naoina/kocha"
	"hoge/app/controllers"
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
