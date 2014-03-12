---
layout: docs
root: ..
title: Model
subnav:
-
  name: Basics
  path: Basics
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
