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
Basically, route and Controller is paired, so never run any Controller if not routing.

## Route definition <a id="Route-definition"></a>

The routes define in `config/routes.go`.

```go
package config

import (
    "github.com/naoina/kocha"
    "your/app/controllers"
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
```

format:

```go
{
    Name:       "root",
    Path:       "/",
    Controller: controllers.Root{},
}
```

`Name` field is a name of route. it use for a reverse routing. ([url]({{ site.godoc }}#TemplateFuncs) function in template, [Reverse]({{ site.godoc }}#Reverse) in Go code)
`Path` field is a routing path. Kocha will be routed to Controller when request path matches `Path`.
For example, If route defined the following:

```go
{
    Name:       "myroom"
    Path:       "/myroom"
    Controller: controllers.Myroom{},
}
```

And when request is `GET /myroom`, it will be routed to *controllers.Myroom.Get* method.
Also when request is `POST /myroom`, it will be routed to *controllers.Myroom.Post* method.
Similarly, for each request, `PUT` to *Put*, `DELETE` to *Delete*, `HEAD` to *Head* and `PATCH` to *Patch* are routed to those methods respectively.

Finally, `Controller` field is an instance of Controller. See [Controller]({{ page.root }}/docs/controller.html) document for more details.

## Route parameter <a id="Route-parameter"></a>

A routing path can specifies the parameters.
They parameters will be validate in boot time.

A route parameter must be started with "**:**" or "__*__". Normally, use "**:**" except you want to get the path parameter.

For example:

```go
Path: "/:name"
```

This is a routing definition that it includes `:name` route parameter.
If *Controller.Get* of this route defined as follows:

```go
func (c *Root) Get(name string) kocha.Result {
    ......
}
```

`:name` route parameter matches any string. (but "**/**" is not included)
e.g. it matches `/alice`, but does not match with `/alice/1`.
Then matched value (`alice` in this example) will be passed to Controller's method as **name** argument.

Also multiple parameters can be specified.
For example,

Route:

```go
Path: "/:id/:name"
```

Controller:

```go
func (c *Root) Get(id int, name string) kocha.Result {
    ......
}
```

Above example matches all of `/1/alice`, `/10/alice`, `/2/bob` and etc.
However, it does not match with `/str/alice` because `:id` route parameter is defined as type *int* in arguments of Controller's method.

Supported parameter types:

* string
* int
* \*url.URL (See [Path parameter](#Path-parameter))

### Path parameter <a id="Path-parameter"></a>

When route parameter starts with "__*__", it matches with word characters, "**.**", "**-**" and "**/**". In regular expression, it is `[\w-/.]+`.

For example,

Route:

```go
Path: "/*path"
```

Controller:

```go
import "net/url"

func (c *Root) Get(path *url.URL) kocha.Result {
    ......
}
```

If the request to the above example is `GET /path/to/static.png`, `path.Path` of *Controller.Get* will be the `path/to/static.png`.
