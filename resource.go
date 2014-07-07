package kocha

// ResourceSet represents a set of pre-loaded resources.
type ResourceSet map[string]interface{}

// Add adds pre-loaded resource.
func (rs *ResourceSet) Add(name string, data interface{}) {
	if *rs == nil {
		*rs = ResourceSet{}
	}
	(*rs)[name] = data
}

// Get gets pre-loaded resource by name.
func (rs ResourceSet) Get(name string) interface{} {
	return rs[name]
}
