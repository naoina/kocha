package kocha

type FixtureRootTestCtrl struct {
	*Controller
}

func (c *FixtureRootTestCtrl) Get() Result {
	return c.Render()
}
