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

A Controller is a layer of handling request.

## Organization <a id="Organization"></a>

```
.
`-- app
    `-- controllers
        `-- root.go`
```

## Basics <a id="Basics"></a>

You can use `kocha` command line tool for create a new controller as following:

    kocha g controller NAME

Where "**NAME**" is the controller name.

NOTE: The following command is same as above because `g` is aliased to `generate` subcommand:

    kocha generate controller NAME

The above command also generates template file into `app/views/` directory and adds route into `config/routes.go` automatically.

## Render <a id="Render"></a>

Kocha provides some renderer for various purpose.

### Render *([godoc]({{ site.godoc }}#Controller.Render))*

Render a template that collect from a template directory in boot time (Usually, *app/views*).

```go
func (c *Root) Get() kocha.Result {
    return c.Render()
}
```

By default, template type is `text/html`. It responds `app/views/root.html`. (Where *root* in *root.html* is the Controller name mapped to snake case)
If ContentType isn't specified, render the file type specific template that it detects by ContentType.

e.g.

```go
func (c *Root) Get() kocha.Result {
    c.Response.ContentType = "application/json"
    return c.Render()
}
```

The above responds `app/views/root.json` template.

Also *Render* can be passed context to [Template.Execute](http://golang.org/pkg/html/template/#Template.Execute):

```go
func (c *Root) Get() kocha.Result {
    return c.Render(kocha.Context{
        "name": "alice",
    })
}
```

### RenderJSON *([godoc]({{ site.godoc }}#Controller.RenderJSON))*

Render a context as JSON. See [json.Marshal](http://golang.org/pkg/encoding/json/#Marshal) for encodes details.

```go
func (c *Root) Get() kocha.Result {
    return c.RenderJSON(kocha.Context{
        "name": "alice",
        "id": 1,
    })
}
```

If you want to your own JSON format, please use [Render](#Render) with ContentType specified to *application/json*.

### RenderXML *([godoc]({{ site.godoc }}#Controller.RenderXML))*

Render a context as XML. See [xml.Marshal](http://golang.org/pkg/encoding/xml/#Marshal) for encodes details.

```go
import "encoding/xml"

func (c *Root) Get() kocha.Result {
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

If you want to your own XML format, please use [Render](#Render) with ContentType specified to *application/xml*.

### RenderText *([godoc]({{ site.godoc }}#Controller.RenderText))*

Render a plain text.

```go
func (c *Root) Get() kocha.Result {
    return c.RenderText("something")
}
```

If you want to templating text, please use [Render](#Render) with ContentType specified to *text/plain*.

### RenderError *([godoc]({{ site.godoc }}#Controller.RenderError))*

Render a template (or returns a status text) with status code.

```go
func (c *Root) Get() kocha.Result {
    return c.RenderError(http.StatusBadRequest)
}
```

See *([Controller.RenderError]({{ site.godoc }}#Controller.RenderError))* for more details.

### SendFile *([godoc]({{ site.godoc }}#Controller.SendFile))*

Render a static file that gets from the static file directory (Usually, *public*).
Or gets from binary included resources (See [True All-in-One binary]({{ page.root }}/docs/deployment.html#True-all-in-one-binary)).

```go
func (c *Root) Get() kocha.Result {
    return c.SendFile("/path/to/file")
}
```

If passed path is absolute, render the content read from the path as it is.
If passed path is relative, First, Try to get the content read from included resources and returns it if successful. Otherwise, static directory adds to the prefix of the path and then will render the content read from the path that.

e.g. The absolute path:

    c.SendFile("/srv/favicon.ico")

The above responds `/srv/favicon.ico`.

The relative path:

    c.SendFile("favicon.ico")

The above responds `public/favicon.ico`.

### Redirect *([godoc]({{ site.godoc }}#Controller.Redirect))*

A shorthand for redirect of both *Http.StatusMovedPermanently (301)* and *http.StatusFound (302)*.

Using *Redirect* renderer:

```go
func (c *Root) Get() kocha.Result {
    // MovedPermanently if second argument is true.
    return c.Redirect("/path/to/redirect", false)
}
```

is same as below:

```go
func (c *Root) Get() kocha.Result {
    c.Response.StatusCode = http.StatusFound
    c.Response.Header().Set("Location", "/path/to/redirect")
    return c.RenderText("")
}
```

## Parameter <a id="Parameter"></a>

Controller can take the routing parameters from arguments.

If route defined as following in `config/routes.go`:

```go
Path: "/:id"
```

Controller can take the parameter as argument of **id** of type *int*:

```go
func (c *Root) Get(id int) kocha.Result {
    return c.Render()
}
```

Note that both name (**id** in this case) MUST be identical between argument name of Controller and routing parameter.

Of course, it can take the different multiple types:

```go
Path: "/:id/:name"
```

```go
func (c *Root) Get(id int, name string) kocha.Result {
    return c.Render()
}
```

For more details to definition of route parameters, see [Route parameter]({{ page.root }}/docs/routing.html#Route-parameter).

## Built-in controller <a id="Built-in-controller"></a>

* [StaticServe]({{ site.godoc }}#StaticServe)
