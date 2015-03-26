# Kocha v0.7.0

This release contains the incompatible changes with previous releases.
Also the feature freeze until 1.0.0 release.

## New features:

* request: Add Context.Request.IsXHR
* render: Add Context.Format
* template: Add `join` template func
* template: Add `flash` template func
* template: Template action delimiters now can be changed
* log: Add RawFormatter
* misc: Add ErrorWithLine

## Incompatible changes:

* cli: Move to `cmd` directory
* template: `{{define "content"}}` on each templates are no longer required
* template: Suffix of template file changed to .tmpl
* render: kocha.Render* back to kocha.Context.Render* and signatures are changed
* kocha: Rename SettingEnv to Getenv
* middleware: Several features are now implemented as the middlewares
* middleware: Change interface signature

## Other changes:

* log: Output to console will be coloring
* Some bugfix

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
