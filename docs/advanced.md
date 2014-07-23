---
layout: docs
root: ..
title: Advanced
subnav:
-
  name: Using as http.Handler
  path: Using-as-http-Handler
---

# Advanced <a id="Advanced"></a>

## Using as http.Handler <a id="Using-as-http-Handler"></a>

Kocha app can also be used as [http.Handler](http://golang.org/pkg/net/http/#Handler). Use [kocha.New]({{ site.godoc }}/#New) instead of [kocha.Run]({{ site.godoc }}/#Run).

In `main.go`:

```go
if err := kocha.Run(config.AppConfig); err != nil {
    panic(err)
}
```

modifies to:

```go
app, err := kocha.New(config.AppConfig)
if err != nil {
    panic(err)
}
log.Fatal(http.ListenAndServe(":9100", app))
```
