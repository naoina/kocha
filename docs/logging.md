---
layout: docs
root: ..
title: Logging
subnav:
-
  name: Basics
  path: Basics
-
  name: Log levels
  path: Log-levels
-
  name: Set loggers
  path: Set-loggers
-
  name: Built-in loggers
  path: Built-in-loggers
---

# Logging <a id="Logging"></a>

**NOTE: Logging of Kocha is experimental features. Maybe they will be changed in future.**

## Basics <a id="Basics"></a>

Kocha provides global logger.

Usage:

```go
kocha.Log.Info("this is a %s", variable)
```

Above example, output is in the *Info* log level.

## Log levels <a id="Log-levels"></a>

* `kocha.Log.Debug`
* `kocha.Log.Info`
* `kocha.Log.Warn`
* `kocha.Log.Error`

To set the loggers for these log levels, set the `AppConfig.Logger` in `config/app.go`.

## Set loggers <a id="Set-loggers"></a>

For example:

```go
AppConfig.Logger = &kocha.Logger{
    DEBUG: kocha.Loggers{kocha.ConsoleLogger(-1)},
    INFO:  kocha.Loggers{kocha.ConsoleLogger(-1)},
    WARN:  kocha.Loggers{kocha.ConsoleLogger(-1)},
    ERROR: kocha.Loggers{kocha.ConsoleLogger(-1)},
}
```

Loggers set to `ConsoleLogger` in above example. Also loggers use the prefix flags of [log](http://golang.org/pkg/log/#pkg-constants) package.
If you want to use the default flags, specify the `-1`.
The default flag is `Ldate | Ltime`.

### Set the multiple loggers

You can register multiple loggers, such as the following:

```go
kocha.Loggers{
    kocha.ConsoleLogger(-1),
    kocha.FileLogger("path/to/logfile", -1),
}
```

This is output to all registered loggers when use the logger of that log level.

## Built-in loggers <a id="Built-in-loggers"></a>

Kocha provides some common loggers.

### ConsoleLogger *([godoc]({{ site.godoc }}#ConsoleLogger))*

Output to `os.Stdout`.

### FileLogger *([godoc]({{ site.godoc }}#FileLogger))*

Output to specified file.

### NullLogger *([godoc]({{ site.godoc }}#NullLogger))*

This is dummy logger that it doesn't output.
