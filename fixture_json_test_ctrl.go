package kocha

type FixtureJsonTestCtrl struct {
	Controller
}

func (c *FixtureJsonTestCtrl) Get() Result {
	c.Response.ContentType = "application/json"
	return c.Render()
}
