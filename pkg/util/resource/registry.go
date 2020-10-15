package resource

type Registry struct {
	Descriptors []Descriptor
}

func (r *Registry) Clone() *Registry {
	r2 := &Registry{}
	r2.Descriptors = make([]Descriptor, len(r.Descriptors))
	copy(r2.Descriptors, r.Descriptors)
	return r2
}

func (r *Registry) Register(desc Descriptor) Descriptor {
	r.Descriptors = append(r.Descriptors, desc)
	return desc
}

var DefaultRegistry = &Registry{}

func RegisterResource(desc Descriptor) Descriptor {
	return DefaultRegistry.Register(desc)
}
