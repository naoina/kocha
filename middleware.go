package kocha

type Middleware interface {
	Before(c *Controller)
	After(c *Controller)
}

var (
	DefaultMiddlewares = []Middleware{
		&ResponseContentTypeMiddleware{},
	}
)

type ResponseContentTypeMiddleware struct{}

func (m *ResponseContentTypeMiddleware) Before(c *Controller) {
	// do nothing
}

func (m *ResponseContentTypeMiddleware) After(c *Controller) {
	res := c.Response
	res.Header().Set("Content-Type", res.ContentType+"; charset=utf-8")
}
