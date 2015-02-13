---
layout: docs
root: ..
title: Controller
subnav:
-
  name: Organization
  path: Organization
-
  name: Basics
  path: Basics
-
  name: Render
  path: Render
-
  name: Parameter
  path: Parameter
-
  name: Built-in controller
  path: Built-in-controller
---

# Controller <a id="Controller"></a>

Controller is the layer of handling request.

## Organization <a id="Organization"></a>

```
.
`-- app
    `-- controller
        `-- root.go`
```

## Basics <a id="Basics"></a>

You can use `kocha` command line tool to create a new controller as follows:

    kocha g controller NAME

Where "**NAME**" is the controller name.

NOTE: The following command is same as above because `g` is aliased to `generate` subcommand:

    kocha generate controller NAME

The above command also generates a template file into `app/view/` directory and adds route into `config/routes.go` automatically.

## Render <a id="Render"></a>

Kocha provides some renderer for various purpose.

### Render *([godoc]({{ site.godoc }}#Render))*

Render a template that collect from a template directory in boot time (Usually, *app/view*).

```go
func (r *Root) GET(c *kocha.Context) error {
    return c.Render(nil)
}
```

By default, template type is `text/html`. It responds `app/view/root.html.tmpl`. (Where *root* in *root.html.tmpl* is the Controller name mapped to snake case)
If ContentType isn't specified, render the file type specific template that it detects by ContentType.

e.g.

```go
func (r *Root) GET(c *kocha.Context) error {
    c.Response.ContentType = "application/json"
    return c.Render(nil)
}
```

The above responds `app/view/root.json` template.

Also *Render* can be passed data to [Template.Execute](http://golang.org/pkg/html/template/#Template.Execute):

```go
func (r *Root) GET(c *kocha.Context) error {
    return c.Render(map[string]interface{}{
        "name": "alice",
    })
}
```

### RenderJSON *([godoc]({{ site.godoc }}#RenderJSON))*

Render context as JSON. See [json.Marshal](http://golang.org/pkg/encoding/json/#Marshal) for encode details.

```go
func (r *Root) GET(c *kocha.Context) error {
    return c.RenderJSON(map[string]interface{}{
        "name": "alice",
        "id": 1,
    })
}
```

If you want to render your own JSON format, please use [Render](#Render) with ContentType specified to *application/json*.

### RenderXML *([godoc]({{ site.godoc }}#RenderXML))*

Render context as XML. See [xml.Marshal](http://golang.org/pkg/encoding/xml/#Marshal) for encode details.

```go
import "encoding/xml"

func (r *Root) GET(c *kocha.Context) error {
    return c.RenderXML(struct {
        XMLName xml.Name `xml:"person"`
        Id      int      `xml:"id"`
        Name    string   `xml:"name"`
    }{
        Id:   1,
        Name: "Alice",
    })
}
```

If you want to render your own XML format, please use [Render](#Render) with ContentType specified to *application/xml*.

### RenderText *([godoc]({{ site.godoc }}#RenderText))*

Render plain text.

```go
func (r *Root) GET(c *kocha.Context) error {
    return c.RenderText("something")
}
```

If you want to templating text, please use [Render](#Render) with ContentType specified to *text/plain*.

### RenderError *([godoc]({{ site.godoc }}#RenderError))*

Render template (or returns status text) with status code.

```go
func (r *Root) GET(c *kocha.Context) error {
    return c.RenderError(nil, http.StatusBadRequest, nil)
}
```

Also you can pass an error to the first argument of `RenderError`.
The passed error will be logging by `c.App.Logger.Error`.

```go
func (r *Root) GET(c *kocha.Context) error {
    if err := DoSomething(); err != nil {
        return c.RenderError(err, 500, nil)
    }
}
```

See *([RenderError]({{ site.godoc }}#RenderError))* for more details.

### SendFile *([godoc]({{ site.godoc }}#SendFile))*

Render a static file that gets from the static file directory (Usually, *public*).
Or gets from binary included resources (See [True All-in-One binary]({{ page.root }}/docs/deployment.html#True-all-in-one-binary)).

```go
func (r *Root) GET(c *kocha.Context) error {
    return c.SendFile("/path/to/file")
}
```

If passed path is absolute, render the content read from the path as it is.
If passed path is relative, First, try to get the content from included resources and returns it if success. Otherwise, The static directory path adds to the prefix of the path and then will render the content read from the path that.

For example, an absolute path:

    c.SendFile("/srv/favicon.ico")

The above responds `/srv/favicon.ico`.

A relative path:

    c.SendFile("favicon.ico")

The above responds `public/favicon.ico`.

### Redirect *([godoc]({{ site.godoc }}#Redirect))*

A shorthand for redirect of both *Http.StatusMovedPermanently (301)* and *http.StatusFound (302)*.

Using *Redirect* renderer:

```go
func (r *Root) GET(c *kocha.Context) error {
    // MovedPermanently if second argument is true.
    return c.Redirect("/path/to/redirect", false)
}
```

is same as below:

```go
func (r *Root) GET(c *kocha.Context) error {
    c.Response.StatusCode = http.StatusFound
    c.Response.Header().Set("Location", "/path/to/redirect")
    return c.RenderText("")
}
```

## Parameter <a id="Parameter"></a>

Controller can take the routing parameters.

If route defined as following in `config/routes.go`:

```go
Path: "/:id"
```

Controller can take the parameter by [Context]({{ site.godoc }}#Context).[Params]({{ site.godoc }}#Params).

```go
func (r *Root) GET(c *kocha.Context) error {
    id := c.Params.Get("id")
    return c.Render(map[string]interface{}{
        "id": id,
    })
}
```

The parameter that taken out is a `string` type. If you want other types, you can convert to other types using such as [strconv](http://golang.org/pkg/strconv/).

Of course, it can take the multiple parameters:

```go
Path: "/:id/:name"
```

```go
func (r *Root) GET(c *kocha.Context) error {
    return c.Render(map[string]interface{}{
        "id": c.Params.Get("id"),
        "name": c.Params.Get("name"),
    })
}
```

For more details of definition of route parameters, see [Route parameter]({{ page.root }}/docs/routing.html#Route-parameter).

## Built-in controller <a id="Built-in-controller"></a>

* [StaticServe]({{ site.godoc }}#StaticServe)
* [ErrorController]({{ site.godoc }}#ErrorController)
