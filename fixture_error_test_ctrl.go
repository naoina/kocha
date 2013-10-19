package kocha

type FixtureErrorTestCtrl struct {
	Controller
}

func (c *FixtureErrorTestCtrl) Get() Result {
	panic("panic test")
	return c.Render()
}
