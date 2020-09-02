package template

type Registry struct {
	items map[string]T
}

func NewRegistry() *Registry {
	return &Registry{items: make(map[string]T)}
}

func (r *Registry) Clone() *Registry {
	items := make(map[string]T)
	for k, v := range r.items {
		items[k] = v
	}
	return &Registry{items: items}
}

func (r *Registry) Register(item T) {
	r.items[item.Type] = item
}

func (r *Registry) Lookup(itemType string) (item T, ok bool) {
	item, ok = r.items[itemType]
	return
}

var DefaultRegistry = NewRegistry()

func Register(item T) T {
	DefaultRegistry.Register(item)
	return item
}
