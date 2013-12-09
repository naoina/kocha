---
layout: docs
root: ..
title: Deployment
subnav:
-
  name: Build and deploy
  path: Build-and-deploy
-
  name: True All-in-One binary
  path: True-all-in-one-binary
-
  name: Graceful restart
  path: Graceful-restart
---

# Deployment <a id="Deployment"></a>

## Build and deploy <a id="Build-and-deploy"></a>

Use Kocha command line tool:

    kocha build prod

Or use `go build`:

    go build -o appname prod.go

These commands are generate a binary of Kocha app to `$GOPATH/appname/appname`.

If using `kocha build`:

    cd $GOPATH/appname
    kocha build prod
    rsync -avz appname public targethost:/path/to/appdir/

If using `go build`:

    cd $GOPATH/appname
    go build -o appname prod.go
    rsync -avz appname public app targethost:/path/to/appdir/

## True All-in-One binary <a id="True-all-in-one-binary"></a>

A generated binary by `kocha build prod` (and `go build`) is not includes the static files. (Usually, static files are in `public` directory)
If you want to generate a binary that static files included, use the following command:

    kocha build -a prod

Deployment of True All-in-One binary is very simple. A binary copy to server and restart it.
You don't have to other files copy to server.

## Graceful restart <a id="Graceful-restart"></a>

Kocha app can also *Graceful restart*. (aka *Hot reload*)
Send *SIGHUP* signal to your Kocha app such as using `kill -HUP` command in order to do it.

Sequence:

1. App receive *SIGHUP*
1. Run a new app process and start new requests acceptance
1. Wait the end of accepted requests in old app process
1. exit the old app process
