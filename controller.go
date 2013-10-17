package kocha

import (
	"log"
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
		log.Panic("too many arguments")
	}
	t := appConfig.TemplateSet.Get(appConfig.AppName, c.Name, "html")
	if t == nil {
		log.Panicf("no such template: %s", appConfig.TemplateSet.Ident(appConfig.AppName, c.Name, "html"))
	}
	return &ResultTemplate{
		Template: t,
		Context:  ctx,
	}
}
