package kocha

import (
	"bytes"
	"io"
)

var (
	// All pre-loaded resources
	includedResources = make(map[string]*resource)
)

// AddResource adds pre-loaded resource.
func AddResource(name, data string) {
	includedResources[name] = &resource{[]byte(data)}
}

// resource is represents a pre-loaded resource.
type resource struct {
	// Data of resource.
	data []byte
}

// Open returns io.ReadSeeker of resource data.
func (r *resource) Open() io.ReadSeeker {
	return bytes.NewReader(r.data)
}
