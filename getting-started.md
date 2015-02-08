---
layout: getting-started
root: .
title: Getting started
subnav:
-
  name: Requirements
  path: Requirements
-
  name: Installation
  path: Installation
-
  name: Create a new application
  path: Create-a-new-application
-
  name: Run the application
  path: Run-the-application
-
  name: Write a request handler
  path: Write-a-request-handler
-
  name: Edit a controller
  path: Edit-a-controller
-
  name: Routing parameter
  path: Routing-parameter
-
  name: Ending
  path: Ending
---

# Getting started <a id="Getting-started"></a>

## Requirements <a id="Requirements"></a>

* Go 1.3 or later (http://golang.org)

## Installation <a id="Installation"></a>

Install the Kocha framework:

    go get -u github.com/naoina/kocha

And command line tool:

    go get -u github.com/naoina/kocha/cmd/...

NOTE: If you want to use a specific Go version and/or specific GOPATH, please use a Go version manager such as [gvm](https://github.com/moovweb/gvm).

## Create a new application <a id="Create-a-new-application"></a>

To create a new Kocha application, run the following command:

    kocha new myapp

Where "**myapp**" is the application name.
Above command will create a skeleton files of app to `$GOPATH/myapp`.

## Run the application <a id="Run-the-application"></a>

Change directory:

    cd $GOPATH/myapp

And run the application:

    kocha run

Then, open http://127.0.0.1:9100/ in your Browser.
Do you see the welcome page?

![fig1]({{ page.root }}/images/fig1.png)

Congratulation!
You've finished the first step of the development of the Kocha app.



## Write a request handler <a id="Write-a-request-handler"></a>

To create a new controller, run the following command:

    kocha g controller myroom

Where "**myroom**" is the controller name.

NOTE: The following command is same as above because `g` is aliased to `generate` subcommand:

    kocha generate controller myroom

Also `generate` subcommand adds route into `config/routes.go` automatically.

## Edit a controller <a id="Edit-a-controller"></a>

So let's edit as follows.

In `app/controller/myroom.go`, edit to:

{% raw %}
```go
package controller

import (
    "github.com/naoina/kocha"
)

type Myroom struct {
    *kocha.DefaultController
}

func (c *Myroom) Get(c *kocha.Context) error {
    return c.Render(map[string]interface{}{
        "name": "Alice",
    })
}
```
{% endraw %}

In `app/view/myroom.html.tmpl`, edit to:

{% raw %}
```html
<h1>This is {{.name}}'s room</h1>
```
{% endraw %}

`kocha run` watch the files and reload when changed.

Please open http://127.0.0.1:9100/myroom in your Browser.

![fig2]({{ page.root }}/images/fig2.png)

You should see the changes that you have made.

Kocha uses Go's [html/template](http://golang.org/pkg/html/template/). See [Template]({{ page.root }}/docs/template.html) for more information.



## Routing parameter <a id="Routing-parameter"></a>

Routing of Kocha can get the parameter from the requested URL path.
So let's do it.

First, In `config/routes.go`, edit:

{% raw %}
```go
Path:       "/myroom",
```
{% endraw %}

to

{% raw %}
```go
Path:       "/myroom/:name",
```
{% endraw %}

Second, In `app/controller/myroom.go`, edit:

{% raw %}
```go
func (c *Myroom) Get(c *kocha.Context) error {
    return c.Render(map[string]interface{}{
        "name": "Alice",
    })
}
```
{% endraw %}

to

{% raw %}
```go
func (c *Myroom) Get(c *kocha.Context) error {
    return c.Render(map[string]interface{}{
        "name": c.Params.Get("name"),
    })
}
```
{% endraw %}

Finally, Let's see http://127.0.0.1:9100/myroom/bob in your Browser as usual.

![fig3]({{ page.root }}/images/fig3.png)

Kocha's routing is more powerful. See [Routing]({{ page.root }}/docs/routing.html) for more information.

## Ending <a id="Ending"></a>

These are a part of Kocha.
Are you more interested in Kocha? OK, see the [Docs]({{ page.root }}/docs/)!
