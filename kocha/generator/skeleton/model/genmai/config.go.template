package db

import (
	"fmt"
	"path/filepath"

	"github.com/naoina/genmai"
	"github.com/naoina/kocha"

	_ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var DatabaseMap = kocha.DatabaseMap{
	"default": {
		Driver: kocha.SettingEnv("KOCHA_DB_DRIVER", "sqlite3"),
		DSN:    kocha.SettingEnv("KOCHA_DB_DSN", filepath.Join("db", "db.sqlite3")),
	},
}

var dbMap = make(map[string]*genmai.DB)

func Get(name string) *genmai.DB {
	return dbMap[name]
}

func init() {
	for name, dbconf := range DatabaseMap {
		var d genmai.Dialect
		switch dbconf.Driver {
		case "mysql":
			d = &genmai.MySQLDialect{}
		case "postgres":
			d = &genmai.PostgresDialect{}
		case "sqlite3":
			d = &genmai.SQLite3Dialect{}
		default:
			panic(fmt.Errorf("kocha: genmai: unsupported driver type: %v", dbconf.Driver))
		}
		db, err := genmai.New(d, dbconf.DSN)
		if err != nil {
			panic(err)
		}
		dbMap[name] = db
	}
}
