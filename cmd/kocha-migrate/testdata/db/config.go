package db

import "github.com/naoina/kocha"

var DatabaseMap = kocha.DatabaseMap{
	"default": {
		Driver: "sqlite3",
		DSN:    ":memory:",
	},
}
