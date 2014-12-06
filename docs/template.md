---
layout: docs
root: ..
title: Template
subnav:
-
  name: Organization
  path: Organization
-
  name: Basics
  path: Basics
-
  name: Layout
  path: Layout
-
  name: File types
  path: File-types
-
  name: Built-in functions
  path: Built-in-functions
---

# Template <a id="Template"></a>

Kocha uses standard Go template format that provided by [html/template](http://golang.org/pkg/html/template/).

## Organization <a id="Organization"></a>

```
.
`-- app
    `-- view
        |-- layout
        |   `-- app.html    # A layout file for HTML file type
        `-- root.html       # HTML template file for Root controller
```

## Basics <a id="Basics"></a>

Template is highly related with [Controller]({{ page.root }}/docs/controller.html).

When Controller name is `root`, a template file name **MUST** be `app/view/root.[extension]`.
`[extension]` is `html` by default. (See [File types](#File-types))

Use `html` extension in this example.

app/view/root.html:

{% raw %}
```html
<h1>Welcome to Kocha</h1>
```
{% endraw %}

Output:

{% raw %}
```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Welcome to Kocha</title>
</head>
<body>
  <h1>Welcome to Kocha</h1>
</body>
</html>
```
{% endraw %}

In fact, white-spaces in output are perhaps different from example above.

## Layout <a id="Layout"></a>

Kocha supports template layout and also enabled by default.
The default layout is `app`, it retrieves `app/view/layout/app.[extension]`.
You can change the default layout by `AppConfig.DefaultLayout` in `config/app.go`.

Also multiple layout files are supported.
To use another layout instead of the default layout, set any layout name to `c.Layout` of Context.

For example, layout name set to `sub`, and when templates and Controller are following.

app/view/layout/sub.html:

{% raw %}
```html
<html>
<head></head>
<body>
  {{yield .}}
  <p>This is the sub layout.</p>
</body>
</html>
```
{% endraw %}

In app/controller/root.go:

{% raw %}
```go
func (r *Root) GET(c *kocha.Context) kocha.Result {
    c.Layout = "sub"
    return kocha.Render(c, nil)
}
```
{% endraw %}

app/view/root.html is same as previous.

Output:

{% raw %}
```html
<html>
<head></head>
<body>
  <h1>Welcome to Kocha</h1>
  <p>This is the sub layout.</p>
</body>
</html>
```
{% endraw %}

## File types <a id="File-types"></a>

You can use template file for each file types.

See also [Render]({{ page.root }}/docs/controller.html#Render).

## Built-in functions *([godoc]({{ site.godoc }}#TemplateFuncs))* <a id="Built-in-functions"></a>

Kocha provides various additional template functions such as follows.

### in

Returns the boolean truth of whether the arg1 contains arg2.

For example, when `arr` is a slice of `{"a", "b", "c"}`:

{% raw %}
```
{{if in arr "b"}}
arr have b.
{{endif}}
```
{% endraw %}

Output:

{% raw %}
```
arr have b.
```
{% endraw %}

### url

An alias for [Reverse]({{ site.godoc }}#Reverse).

### nl2br

Convert `"\n"` to `<br>` tags.

Example:

{% raw %}
```
{{nl2br "some\ntext"}}
```
{% endraw %}

Output:

{% raw %}
```
some<br>text
```
{% endraw %}

### raw

Input string outputs it is. It won't be escaped.

Example:

{% raw %}
```
{{raw "some<br>text"}}
```
{% endraw %}

Output:

{% raw %}
```
some<br>text
```
{% endraw %}
