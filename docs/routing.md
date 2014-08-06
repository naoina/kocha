---
layout: docs
root: ..
title: Routing
subnav:
-
  name: Route definition
  path: Route-definition
-
  name: Route parameter
  path: Route-parameter
---

# Routing <a id="Routing"></a>

Routing will bridge the requested path and [Controller]({{ page.root }}/docs/controller.html).
Basically, route and Controller are paired, so never run any Controller if routing is not defined.

## Route definition <a id="Route-definition"></a>

You can define routing in `config/routes.go`.

```go
package config

import (
    "github.com/naoina/kocha"
    "your/app/controller"
)

type RouteTable kocha.RouteTable

var routes = RouteTable{
    {
        Name:       "root",
        Path:       "/",
        Controller: &controller.Root{},
    },
}

func Routes() RouteTable {
    return append(routes, RouteTable{
        {
            Name:       "static",
            Path:       "/*path",
            Controller: &kocha.StaticServe{},
        },
    }...)
}
```

format:

```go
{
    Name:       "root",
    Path:       "/",
    Controller: &controller.Root{},
}
```

`Name` field is name of route. it use for reverse routing. ([url]({{ site.godoc }}#TemplateFuncs) function in template, [Reverse]({{ site.godoc }}#Reverse) in Go code)
`Path` field is routing path. Kocha will be routed to Controller when request path matches `Path`.
For example, If route is defined the following:

```go
{
    Name:       "myroom"
    Path:       "/myroom"
    Controller: &controller.Myroom{},
}
```

And when request is `GET /myroom`, it will be routed to *controller.Myroom.GET* method.
Also when request is `POST /myroom`, it will be routed to *controller.Myroom.POST* method.

## Route parameter <a id="Route-parameter"></a>

Routing path can specify parameters.
They parameters will be validated in boot time.

Route parameter must be started with "**:**" or "__*__". Normally, use "**:**" except you want to get the path parameter.

For example:

```go
Path: "/:name"
```

This is routing definition that it includes `:name` route parameter.
If *Controller.GET* of that route is defined as follows:

```go
func (r *Root) GET(c *kocha.Context) kocha.Result {
    c.Params.Get("name")
    ......
}
```

`:name` route parameter matches any string. (but "**/**" is not included)
For example, it will match `/alice`, but won't match `/alice/1`.
Then matched value (`alice` in this example) will be stored to [Context]({{ site.godoc }}#Context).[Params]({{ site.godoc }}#Params) as **name** key.

Also multiple parameters can be specified.
For example,

Route:

```go
Path: "/:id/:name"
```

Controller:

```go
func (r *Root) GET(c *kocha.Context) kocha.Result {
    c.Params.Get("id")
    c.Params.Get("name")
    ......
}
```

Above example matches all of `/1/alice`, `/10/alice`, `/2/bob`, `/str/alice` and etc.

### Path parameter <a id="Path-parameter"></a>

When route parameter starts with "__*__", it will match all word characters.

For example,

Route:

```go
Path: "/*path"
```

Controller:

```go
func (r *Root) GET(c *kocha.Context) kocha.Result {
    c.Params.Get("path")
    ......
}
```

If `GET /path/to/static.png` requests to the above example, *Context.Params.Get* will return `"path/to/static.png"`.
