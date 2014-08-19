# Kocha v0.6.1

## Changes

* [bugfix] Fix a problem that reloading process of `kocha run' doesn't work

# Kocha v0.6.0

This release is an incompatible with previous releases.

## New features

* feature: add middleware for Flash messaging
* cli: CLI now can append the user-defined subcommands like git
* session: add Get, Set and Del API

## Incompatible changes

* all: names of packages in an application to change to singular name
* log: logger is fully redesigned
* template: Remove `date' template function
* renderer: Move kocha.Context.Render* to kocha.Render*
* context: Change Errors() method to the Errors field
* middleware: Middleware.After will be called in the reverse of the order in which Middleware.Before are called
* controller: controller types are fully redesigned
* controller: remove NewErrorController()
* router: Route.dispatch won't create an instance of Controller for each dispatching
* middleware: processing of ResponseContentTypeMiddleware moves to core

## Other changes

* session: codec.MsgpackHandle won't be created on each call
* all: refactoring
