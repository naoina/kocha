package kocha

import (
	"../../kocha"
)

type FixtureRelpathTestCtrl struct {
	kocha.Controller
}

func (c *FixtureRelpathTestCtrl) Get() kocha.Result {
	return c.Render()
}
