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
  name: Implementing a session store
  path: Implementing-a-session-store
---

# Session <a id="Session"></a>

Session is a simple key-value storage.
It will save and load for every request.

## Basics <a id="Basics"></a>

For example, do a following in order to set the data in the session:

```go
func (c *Root) Get() kocha.Result {
    c.Session["name"] = "alice"
    ......
}
```

And load the data from session:

```go
func (c *Root) Get() kocha.Result {
    name := c.Session["name"]
    ......
}
```

To delete a data from session, use the `delete` built-in function:

```go
func (c *Root) Get() kocha.Result {
    c.Session["name"] = "alice"
    name := c.Session["name"] // returns "alice"
    delete(c.Session, "name")
    name = c.Session["name"] // returns ""
    ......
}
```

Also Session has the `Clear` method that clear the all data from session.

```go
func (c *Root) Get() kocha.Result {
    c.Session["name"] = "alice"
    c.Session["id"] = "1"
    l := len(c.Session) // returns 2
    c.Session.Clear()
    l = len(c.Session) // returns 0
    ......
}
```

Actually, session is a string map `map[string]string`.
Therefore, if you want to save the non-string data, please serialize the data to string on their own.

## Session store <a id="Session-store"></a>

Now currently, Kocha provides Cookie store only.
Cookie store saves session data to a client-side cookie with encrypted.
It's independent of the other system/server, but expiry date isn't fully controllable.
If you want to do it, please implements a session store that use server-side storage such as a memcached or database. See [Implementing a session store](#Implementing-a-session-store).

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

If you want to do not use the session, please remove the `kocha.SessionMiddleware` from `AppConfig.Middlewares` in `config/[env]/app.go`.

## Implementing a session store <a id="Implementing-a-session-store"></a>

If you want other session store that not provided by Kocha, you can implements your own session store.

1. Implements the [SessionStore]({{ site.godoc }}#SessionStore) interface.
1. It specify to the `AppConfig.Store` in `config/app.go`.

Also, see source of [SessionCookieStore]({{ site.godoc }}#SessionCookieStore) as example.
