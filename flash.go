package kocha

// Flash represents a container of flash messages.
// Flash is for the one-time messaging between requests. It useful for
// implementing the Post/Redirect/Get pattern.
type Flash map[string]FlashData

// Get gets a value associated with the given key.
// If there is the no value associated with the key, Get returns "".
func (f Flash) Get(key string) string {
	if f == nil {
		return ""
	}
	data, exists := f[key]
	if !exists {
		return ""
	}
	data.Loaded = true
	f[key] = data
	return data.Data
}

// Set sets the value associated with key.
// It replaces the existing value associated with key.
func (f Flash) Set(key, value string) {
	if f == nil {
		return
	}
	data := f[key]
	data.Loaded = false
	data.Data = value
	f[key] = data
}

// Len returns a length of the dataset.
func (f Flash) Len() int {
	return len(f)
}

// deleteLoaded delete the loaded data.
func (f Flash) deleteLoaded() {
	for k, v := range f {
		if v.Loaded {
			delete(f, k)
		}
	}
}

// FlashData represents a data of flash messages.
type FlashData struct {
	Data   string // flash message.
	Loaded bool   // whether the message was loaded.
}
