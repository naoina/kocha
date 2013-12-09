---
layout: docs
root: ..
title: Docs
subnav:
-
  name: Organization
  path: Organization
---

# Introduction <a id="Introduction"></a>

Kocha is a convenient web application framework.
In order to learn basic usage of Kocha, read [Getting started]({{ page.root }}/getting-started.html).

## Organization <a id="Organization"></a>

```
.
|-- app
|   |-- controllers
|   |   `-- root.go
|   `-- views
|       |-- layouts
|       |   `-- app.html
|       `-- root.html
|-- config
|   |-- app.go
|   |-- dev
|   |   `-- app.go
|   |-- prod
|   |   `-- app.go
|   `-- routes.go
|-- dev.go
|-- prod.go
`-- public
    `-- robots.txt
```

`dev.go` and `prod.go` are entry point for each `dev` and `prod` environment. Those will be referred to `[env]` in this Docs.
`config/app.go` is a common configuration, and `config/[env]/app.go` are specific configuration for each environment.
