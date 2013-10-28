package kocha

type config map[string]interface{}

var configs = make(map[string]config)

func Config(appName string) config {
	if c, ok := configs[appName]; ok {
		return c
	}
	c := make(config)
	configs[appName] = c
	return c
}

func (c config) Get(name string) (result interface{}, ok bool) {
	result, ok = c[name]
	return result, ok
}

func (c config) Set(name string, value interface{}) {
	c[name] = value
}

func (c config) Int(name string) (result int, ok bool) {
	if v, ok := c[name]; ok {
		if result, ok := v.(int); ok {
			return result, true
		}
	}
	return 0, false
}

func (c config) IntDefault(name string, def int) (result int) {
	if result, ok := c.Int(name); ok {
		return result
	}
	return def
}

func (c config) String(name string) (result string, ok bool) {
	if v, ok := c[name]; ok {
		if result, ok := v.(string); ok {
			return result, true
		}
	}
	return "", false
}

func (c config) StringDefault(name string, def string) (result string) {
	if result, ok := c.String(name); ok {
		return result
	}
	return def
}

func (c config) Bool(name string) (result bool, ok bool) {
	if v, ok := c[name]; ok {
		if result, ok := v.(bool); ok {
			return result, true
		}
	}
	return false, false
}

func (c config) BoolDefault(name string, def bool) (result bool) {
	if result, ok := c.Bool(name); ok {
		return result
	}
	return def
}
