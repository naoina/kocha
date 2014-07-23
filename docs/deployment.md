---
layout: docs
root: ..
title: Deployment
subnav:
-
  name: Build the application
  path: Build-the-application
-
  name: Deploy the application
  path: Deploy-the-application
-
  name: Migration
  path: Migration
-
  name: True All-In-One binary
  path: True-All-In-One-binary
-
  name: Graceful restart
  path: Graceful-restart
---

# Deployment <a id="Deployment"></a>

## Build the application<a id="Build-the-application"></a>

Use Kocha command line tool:

    kocha build

Or use `go build`:

    go build -o appname

These commands are generate a binary of Kocha app to `$GOPATH/appname/appname`.

## Deploy the application<a id="Deploy-the-application"></a>

You can use any file transfer tool such as `rsync` in order to deploy the your application.

    rsync -avz appname public targethost:/path/to/appdir/

## Migration <a id="Migration"></a>

In development environment, use `kocha migrate` command for migration.
You can do migration using the built your application in the same way.

For forward migration:

    appname migrate up

For backward migration:

    appname migrate down

Where **appname** is your application name.
Please see [Migration]({{ page.root }}/docs/model.html#Migration) for more details.

## True All-In-One binary <a id="True-All-In-One-binary"></a>

A generated binary by `kocha build` (or `go build`) doesn't include static files. (Usually, static files are in `public` directory)
If you want to generate a binary which include static files, use the following command:

    kocha build -a

Deployment of True All-in-One binary is very simple. You just transfer that binary to the server and restart it.
You don't have to transfer other files to server.

## Graceful restart <a id="Graceful-restart"></a>

Kocha app can be *Graceful restart* (aka *Hot reload*) if you are using `kocha.Run`.
Send *SIGHUP* signal to your Kocha app such as using `kill -HUP` command in order to do it.

Sequence:

1. App receive *SIGHUP*
1. Run a new app process and starts new requests acceptance
1. Wait the end of accepted requests in old app process
1. exit the old app process

In fact, `kocha.Run` is using [github.com/naoina/miyabi](https://github.com/naoina/miyabi) in order to graceful stop/restart.
If you want to change the signals to graceful stop/restart, please see document of [github.com/naoina/miyabi](https://github.com/naoina/miyabi).
