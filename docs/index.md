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
|   `-- routes.go
|-- main.go
`-- public
    `-- robots.txt
```

`main.go` is entry point and `config/app.go` is a main configuration file.
They are plain Go source files, so you can be customized if you want.
