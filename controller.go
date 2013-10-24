package kocha

import (
	"errors"
)

type Controller struct {
	Name     string
	Request  *Request
	Response *Response
}

type Context map[string]interface{}

func (c *Controller) Render(context ...Context) Result {
	var ctx Context
	switch len(context) {
	case 0: // do nothing
	case 1:
		ctx = context[0]
	default: // > 1
		panic(errors.New("too many arguments"))
	}
	t := appConfig.TemplateSet.Get(appConfig.AppName, c.Name, "html")
	if t == nil {
		panic(errors.New("no such template: " + appConfig.TemplateSet.Ident(appConfig.AppName, c.Name, "html")))
	}
	return &ResultTemplate{
		Template: t,
		Context:  ctx,
	}
}

func (c *Controller) RenderJSON(context interface{}) Result {
	return &ResultJSON{
		Context: context,
	}
}
