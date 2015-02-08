---
layout: docs
root: ..
title: Session
subnav:
-
  name: Basics
  path: Basics
-
  name: Session store
  path: Session-store
-
  name: Configuration
  path: Configuration
-
  name: Implement the session store
  path: Implement-the-session-store
---

# Session <a id="Session"></a>

Session is simple key-value storage for each user/client.
This feature is provided by [SessionMiddleware]({{ page.root }}/docs/middleware.html#SessionMiddleware)

## Basics <a id="Basics"></a>

For example, do the following in order to set the data into the session:

```go
func (r *Root) GET(c *kocha.Context) error {
    c.Session.Set("name", "alice")
    ......
}
```

And load the data from session:

```go
func (r *Root) GET(c *kocha.Context) error {
    name := c.Session.Get("name")
    ......
}
```

To delete the data from session, use the `delete` built-in function:

```go
func (r *Root) GET(c *kocha.Context) error {
    c.Session.Set("name", "alice")
    name := c.Session.Get("name") // returns "alice".
    c.Session.Del("name")
    name = c.Session.Get("name") // returns "".
    ......
}
```

Also Session has `Clear` method that clear all data from the session.

```go
func (r *Root) GET(c *kocha.Context) error {
    c.Session.Set("name", "alice")
    c.Session.Set("id", "1")
    l := len(c.Session) // returns 2
    c.Session.Clear()
    l = len(c.Session) // returns 0
    ......
}
```

Actually, session is string map `map[string]string`.
Therefore, if you want to save non-string data, please serialize the data to string on their own.

## Session store <a id="Session-store"></a>

Now currently, Kocha provides Cookie store only.
Cookie store saves session data to a client-side cookie with encrypted.
It's independent of the other system/server, but expiry date isn't fully controllable.
If you want to do it, please implements session store that use server-side storage such as memcached or database. See [Implement the session store](#Implement-the-session-store).

## Configuration <a id="Configuration"></a>

General session settings are `AppConfig.Session` in `config/app.go`.

{% raw %}
```go
// Session settings
Session: kocha.SessionConfig{
    Name: "appname_session",
    Store: &kocha.SessionCookieStore{
        // AUTO-GENERATED Random keys. DO NOT EDIT.
        SecretKey:  "......",
        SigningKey: "......",
    },

    // Expiration of session cookie, in seconds, from now.
    // Persistent if -1, For not specify, set 0.
    CookieExpires: time.Duration(90) * time.Hour * 24,

    // Expiration of session data, in seconds, from now.
    // Perssitent if -1, For not specify, set 0.
    SessionExpires: time.Duration(90) * time.Hour * 24,
    HttpOnly:       false,
},
```
{% endraw %}

If you don't want to use the session, please remove `kocha.SessionMiddleware` from `AppConfig.Middlewares` in `config/app.go`.

## Implement the session store <a id="Implement-the-session-store"></a>

If you want other session store that not provided by Kocha, you can implement your own session store.

1. Implements the [SessionStore]({{ site.godoc }}#SessionStore) interface.
1. It specify to `AppConfig.Store` in `config/app.go`.

Also, see source of [SessionCookieStore]({{ site.godoc }}#SessionCookieStore) for examples.
