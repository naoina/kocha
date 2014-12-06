---
layout: docs
root: ..
title: Logging
subnav:
-
  name: Basics
  path: Basics
-
  name: Configurations
  path: Configurations
-
  name: Log levels
  path: Log-levels
-
  name: Formatter
  path: Formatter
---

# Logging <a id="Logging"></a>

Kocha provides a level-based logger.

## Basics <a id="Basics"></a>

Logging feature is provided by [Application.Logger]({{ site.godoc }}#Application.Logger).

Usage:

In controller:

```go
func (r *Root) GET(c *kocha.Context) kocha.Result {
    c.App.Logger.Info("This is a root")
    return kocha.Render(c, nil)
}
```

## Configurations <a id="Configurations"></a>

Logger configurations is in `config/app.go`.

```go
// Logger settings.
Logger: &kocha.LoggerConfig{
    Writer:    os.Stdout,
    Formatter: &log.LTSVFormatter{},
    Level:     log.INFO,
}
```

The definition of [LoggerConfig]({{ site.godoc }}#LoggerConfig) is following.

```go
// LoggerConfig represents the configuration of the logger.
type LoggerConfig struct {
    Writer    io.Writer     // output destination for the logger.
    Formatter log.Formatter // formatter for log entry.
    Level     log.Level     // log level.
}
```

## Log levels <a id="Log-levels"></a>

The log levels are following.

* log.NONE *([godoc]({{ site.godoc }}/log#NONE))*
* log.DEBUG *([godoc]({{ site.godoc }}/log#DEBUG))*
* log.INFO *([godoc]({{ site.godoc }}/log#INFO))*
* log.WARN *([godoc]({{ site.godoc }}/log#WARN))*
* log.ERROR *([godoc]({{ site.godoc }}/log#ERROR))*
* log.FATAL *([godoc]({{ site.godoc }}/log#FATAL))*
* log.PANIC *([godoc]({{ site.godoc }}/log#PANIC))*

You can set a log level to Logger by some ways below.

* Set log level to [LoggerConfig.Level]({{ site.godoc }}#LoggerConfig.Level) in `config/app.go` configuration file.
* Use [log.Logger.SetLevel]({{ site.godoc }}/log#Logger.SetLevel) API

### Suppress the output by log level

For example, when log level set to *log.INFO*.

```go
Logger.Debug("debug log")
```

It won't be output because log level of [Logger.Debug()]({{ site.godoc }}/log#Logger.Debug) has a log.DEBUG, and lower than log.INFO.

```go
Logger.Info("info log")
Logger.Error("error log")
```

It will be output because [Logger.Info()]({{ site.godoc }}/log#Logger.Info) and [Logger.Error]({{ site.godoc }}/log#Logger.Error) have a log level which equal or upper than log.INFO.

## Formatter <a id="Formatter"></a>

Logger has a formatter that format to specific log format.

### Built-in formatter <a id="Built-in-formatter"></a>

Kocha provides formatter below.

#### LTSVFormatter *([godoc]({{ site.godoc }}/log#LTSVFormatter))*

A formatter of *Labeled Tab-separated Values* (LTSV).
This is the default formatter of Kocha if you haven't specified the formatter.
See http://ltsv.org/ for more details of LTSV.

```text
level:INFO	time:2014-07-30T17:45:40.419347835+09:00	method:GET	protocol:HTTP/1.1	status:200	uri:/
level:INFO	time:2014-07-30T17:45:48.238892704+09:00	method:GET	protocol:HTTP/1.1	status:404	uri:/user
```

### Custom formatter <a id="Custom-formatter"></a>

You can define your own custom formatter.

1. Implements the [log.Format]({{ site.godoc }}/log#Format) interface.
1. It set to `AppConfig.Logger.Formatter` in `config/app.go`.
