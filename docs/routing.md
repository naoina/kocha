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
-
  name: Type validator and parser
  path: Type-validator-and-parser
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

`Name` field is name of route. it use for reverse routing. ([url]({{ site.godoc }}#TemplateFuncs) function in template, [Reverse]({{ site.godoc }}#Reverse) in Go code)
`Path` field is routing path. Kocha will be routed to Controller when request path matches `Path`.
For example, If route is defined the following:

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

Finally, `Controller` field is instance of Controller. See [Controller]({{ page.root }}/docs/controller.html) document for more details.

## Route parameter <a id="Route-parameter"></a>

Routing path can specify parameters.
They parameters will be validated in boot time.

Route parameter must be started with "**:**" or "__*__". Normally, use "**:**" except you want to get the path parameter.

For example:

```go
Path: "/:name"
```

This is routing definition that it includes `:name` route parameter.
If *Controller.Get* of that route is defined as follows:

```go
func (c *Root) Get(name string) kocha.Result {
    ......
}
```

`:name` route parameter matches any string. (but "**/**" is not included)
For example, it will match `/alice`, but won't match `/alice/1`.
Then matched value (`alice` in this example) will be passed to method of Controller as **name** argument.

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
However, it won't match `/str/alice` because `:id` route parameter is defined as type *int* in arguments of method of Controller.

Pre-defined parameter types:

* string
* int
* \*url.URL

You can also override and/or define any types, See [Type validator and parser](#Type-validator-and-parser).

### Path parameter <a id="Path-parameter"></a>

When route parameter starts with "__*__", it will match word characters, "**.**", "**-**" and "**/**". In regular expression, it is `[\w-/.]+`.

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

If `GET /path/to/static.png` requests to the above example, `path.Path` of *Controller.Get* will be `path/to/static.png`.

## Type validator and parser <a id="Type-validator-and-parser"></a>

Type validator is validator of a path parameter for any format. It is used in dispatcher and reverse router.
Type parser that string of path parameter parses to value of type of Golang. It is used in dispatcher.

### Define the TypeValidateParser <a id="Define-the-TypeValidateParser"></a>

Some validator and parser of type parameters (`string`, `int` and `*url.URL`) are pre-defined by Kocha.
If you want validator and parser for any types, you can define them.

1\. You must implement the [TypeValidateParser]({{ site.godoc }}#TypeValidateParser) interface.

```go
type TypeValidateParser interface {
    // Validate returns whether the valid value as any type.
    Validate(v interface{}) bool

    // Parse returns value that parses v as any type.
    Parse(v string) (value interface{}, err error)
}
```

2\. Set the your own `TypeValidateParser` for any type.

```go
SetTypeValidateParser("bool", &YourOwnTypeValidateParser{})
```

#### Example <a id="Example"></a>

In this example, define the own `TypeValidateParser` for `bool` type.

Define the `BoolTypeValidateParser` as follows in `config/routes.go`:

```go
type BoolTypeValidateParser struct{}

func (validateParser *BoolTypeValidateParser) Validate(v interface{}) bool {
    switch t := v.(type) {
    case bool:
        return true
    case int:
        return t == 1 || t == 0
    }
    return false
}

func (validateParser *BoolTypeValidateParser) Parse(s string) (data interface{}, err error) {
    switch s {
    case "true", "1":
        return true, nil
    case "false", "0":
        return false, nil
    }
    return false, fmt.Errorf("invalid")
}

func init() {
    SetTypeValidateParser("bool", &BoolTypeValidateParser{})
    AppConfig.Router = kocha.InitRouter(kocha.RouteTable(Routes()))
}
```

Then route modifies to following:

```go
{
    Name:       "root",
    Path:       "/:b",
    Controller: controllers.Root{},
}
```

And also modifies argument of the `Root` controller:

```go
func (c *Root) Get(b bool) kocha.Result {
    // do something.
}
```

It's completed that definition of the TypeValidateParser for bool type.
You can now access to either `/true`, `/false`, `/1` and `/0`.
