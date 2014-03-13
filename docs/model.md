---
layout: docs
root: ..
title: Model
subnav:
-
  name: Basics
  path: Basics
-
  name: Using ORM
  path: Using-ORM
-
  name: Database configuration
  path: Database-configuration
-
  name: Migration
  path: Migration
---

# Model <a id="Model"></a>

Model is data model for your application.
Usually, model will be mapped to database table by Object-relational mapper (ORM).

## Basics <a id="Basics"></a>

Kocha provides generator of model, so you can use `kocha g model` command to generate a model.

    kocha g model NAME

For example:

    kocha g model user name:string age:int

The above command will generate the following files.

```
.
|-- app
|   `-- models
|       `-- user.go
`-- db
    `-- config.go`  # if not generated yet.
```

`app/models/user.go`:

{% raw %}
```go
package models

import (
    "github.com/naoina/genmai"
)

type User struct {
    Id   int64  `db:"pk" json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`

    genmai.TimeStamp
}

......
```
{% endraw %}

By default, Kocha is using [genmai](https://github.com/naoina/genmai) ORM.

## Using ORM <a id="Using-ORM"></a>

First, you write the following import path to files where you use ORM.

```go
import "appname/db"
```

Second, get an instance of ORM and will use it.

```go
dbinst := db.Get("default")
```

`db.Get` can use in controllers, models as well as other Go files.

Also you can read generated file `db/config.go` for more information.

## Database configuration <a id="Database-configuration"></a>

The configurations of databases are defined in `db/config.go` by default.
You can specify database driver and data source name to use by some environment variables.

`KOCHA_DB_DRIVER` to set the driver name of the database such as "mysql".
`KOCHA_DB_DSN` to set the data source name for the database such as "user:password@/dbname"

Note that these environment variables are used in run-time, not compile-time.

To change to different database configuration, run the your application such as follows:

    KOCHA_DB_DRIVER="mysql" KOCHA_DB_DSN="user:password@/dbname" kocha run

Or if you run the your application built by `kocha build` or `go build`.

    KOCHA_DB_DRIVER="mysql" KOCHA_DB_DSN="user:password@/dbname" appname

This method is inspired by [config](http://12factor.net/config) of [The Twelve-Factor App](http://12factor.net/).

## Migration <a id="Migration"></a>

Kocha also supports database migration.
To generate a migration file, use `kocha g migration` command.

    kocha g migration NAME

For example:

    kocha g migration create_user_table

The above command will generate the following files.

```
.
`-- db
    `-- migrations
        |-- 20140312091159_create_user_table.go
        `-- init.go
```

Where *20140312091159* is the timestamp of when generated that.
It is different each time it is generated.

`db/migrations/20140312091159_create_user_table.go`:

{% raw %}
```go
package migrations

import "github.com/naoina/genmai"

func (m *Migration) Up_20140312091159_CreateUserTable(tx *genmai.DB) {
    // FIXME: Update database schema and/or insert seed data.
}

func (m *Migration) Down_20140312091159_CreateUserTable(tx *genmai.DB) {
    // FIXME: Revert the change done by Up_20140312091159_CreateUserTable.
}
```
{% endraw %}

A generated file has two methods for migration.
The method names start with `Up_` are for forward migration, and `Down_` are for backward migration.
If panic in these methods, Kocha will rollback the transaction of that migration.

You should write migration codes in those methods. Then type the following command to do migration:

    kocha migrate up

If you want to rollback the migration, type the following command:

    kocha migrate down

By default, `kocha migrate up` run the all migrations and `kocha migrate down` rollback the one of the most recent migration.
