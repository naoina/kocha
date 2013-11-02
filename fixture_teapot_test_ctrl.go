package kocha

import (
	"net/http"
)

type FixtureTeapotTestCtrl struct {
	Controller
}

func (c *FixtureTeapotTestCtrl) Get() Result {
	c.Response.StatusCode = http.StatusTeapot
	return c.Render()
}
