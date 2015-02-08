---
layout: docs
root: ..
title: Middleware
subnav:
-
  name: Implement the middleware
  path: Implement-the-middleware
-
  name: Built-in middlewares
  path: Built-in-middlewares
---

# Middleware <a id="Middleware"></a>

A middleware is pre/post processor of a request.

## Implement the middleware <a id="Implement-the-middleware"></a>

1. Implements the [Middleware]({{ site.godoc }}#Middleware) interface.
1. It adds to the `AppConfig.Middlewares` in `config/app.go`.

Middleware interface definition is following:

{% raw %}
```go
type Middleware interface {
	Before(app *Application, c *Context) error
	After(app *Application, c *Context) error
}
```
{% endraw %}

`Before` method will be executed before processing of Controller, and `After` method will be executed after processing of Controller.

## Built-in middlewares <a id="Built-in-middlewares"></a>

Kocha provides some middlewares.

### SessionMiddleware *([godoc]({{ site.godoc }}#SessionMiddleware))*

SessionMiddleware will autosave and autoload a session on around a request processing.

### FlashMiddleware *([godoc]({{ site.godoc }}#FlashMiddleware))* <a id="FlashMiddleware"></a>

FlashMiddleware will provided one-time messaging between the requests (aka flash messages).
This middleware depends on SessionMiddleware.

### RequestLoggingMiddleware *([godoc]({{ site.godoc }}#RequestLoggingMiddleware))*

RequestLoggingMiddleware will output an access log.
