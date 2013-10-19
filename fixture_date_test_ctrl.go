package kocha

type FixtureDateTestCtrl struct {
	Controller
}

func (c *FixtureDateTestCtrl) Get(year, month int, day int, name string) Result {
	return c.Render(Context{
		"year":  year,
		"month": month,
		"day":   day,
		"name":  name,
	})
}
