---
layout: docs
root: ..
title: Middleware
subnav:
-
  name: Implementing a middleware
  path: Implementing-a-middleware
-
  name: Built-in middlewares
  path: Built-in-middlewares
---

# Middleware <a id="Middleware"></a>

A middleware is pre/post processor of a request.

## Implementing a middleware <a id="Implementing-a-middleware"></a>

1. Implements the [Middleware]({{ site.godoc }}#Middleware) interface.
1. It adds to the `AppConfig.Middlewares` in `app/[env]/app.go`.

Middleware interface definition is following:

{% raw %}
```go
type Middleware interface {
	Before(c *Controller)
	After(c *Controller)
}
```
{% endraw %}

`Before` method execute a before processing of Controller, and `After` method execute a after processing of Controller.

## Built-in middlewares <a id="Built-in-middlewares"></a>

Kocha provides some middlewares.

### ResponseContentTypeMiddleware *([godoc]({{ site.godoc }}#ResponseContentTypeMiddleware))*

ResponseContentTypeMiddleware adds *Content-Type* header to response header.
This middleware is enabled by default.

### SessionMiddleware *([godoc]({{ site.godoc }}#SessionMiddleware))*

SessionMiddleware will autosave and autoload a session on around a request processing.

### RequestLoggingMiddleware *([godoc]({{ site.godoc }}#RequestLoggingMiddleware))*

RequestLoggingMiddleware will output the access log. This is for development purposes.
