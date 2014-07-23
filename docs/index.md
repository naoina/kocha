---
layout: docs
root: ..
title: Docs
subnav:
-
  name: Features
  path: Features
-
  name: Requirement
  path: Requirement
-
  name: Organization
  path: Organization
---

# Introduction <a id="Introduction"></a>

Kocha is a convenient web application framework.
In order to learn basic usage of Kocha, see [Getting started]({{ page.root }}/getting-started.html).

## Features <a id="Features"></a>

* Batteries included
* All configurations are in Go's syntax
* [Generate an All-In-One binary]({{ page.root }}/docs/deployment.html#Build-the-application)
* [Compatible with net/http]({{ page.root }}/docs/advanced.html#Using-as-http-Handler)

## Requirement <a id="Requirement"></a>

* Go 1.3 or later (http://golang.org)

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
They are plain Go source files, so you can be customized without learn the own syntax for configuration file.
