---
layout: docs
root: ..
title: Flash
subnav:
-
  name: Basics
  path: Basics
---

# Flash <a id="Flash"></a>

Flash is for the one-time messaging between requests. It useful for
implementing the [Post/Redirect/Get](http://en.wikipedia.org/wiki/Post/Redirect/Get) pattern.
This feature is provided by [FlashMiddleware]({{ page.root }}/docs/middleware.html#FlashMiddleware).

## Basics <a id="Basics"></a>

For example, when controller is below and route `/comment` is set to it,

```go
func (r *Root) GET(c *kocha.Context) kocha.Result {
    name := c.Flash.Get("name")
    msg := c.Flash.Get("msg")
    return kocha.Render(c, kocha.Data{
        "name": name,
        "msg": msg,
    })
}

func (r *Root) POST(c *kocha.Context) kocha.Result {
    c.Flash.Set("name", "alice")
    c.Flash.Set("msg", "your comment has been posted!")
    // do something...
    return kocha.Redirect(c, "/comment", false)
}
```

Sequence flow:

1. A client requests `GET /comment`. At this point, `name` and `msg` flash messages are empty.
1. A client requests `POST /comment` such as through a form submission.
1. POST handler will set `name` and `msg` flash messages, then returns a redirect response.
1. A client requests `GET /comment` according to a redirect response. `name` and `msg` have been set by the previous POST handler.
1. When client will request `GET /comment` again, `name` and `msg` are empty because they have been get in the previous GET handler.
