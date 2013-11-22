package kocha

type FixtureUserTestCtrl struct {
	*Controller
}

func (c *FixtureUserTestCtrl) Get(id int) Result {
	return c.Render(Context{
		"id": id,
	})
}
