package graphqlutil

type Lazy struct {
	init  func() (any, error)
	value any
	err   error
}

func NewLazy(init func() (any, error)) *Lazy {
	return &Lazy{init: init}
}

func NewLazyValue(value any) *Lazy {
	switch value := value.(type) {
	case *Lazy:
		return value
	case func() (any, error):
		return NewLazy(value)
	default:
		return &Lazy{value: value}
	}
}

func NewLazyError(err error) *Lazy {
	return &Lazy{err: err}
}

func (l *Lazy) Value() (any, error) {
	if l.init != nil {
		l.value, l.err = l.init()
		l.init = nil
	}

	// Recursively evaluate Lazy
	for {
		lazy, ok := l.value.(*Lazy)
		if !ok {
			break
		}
		l.value, l.err = lazy.Value()
		if l.err != nil {
			return nil, l.err
		}
	}

	// Evaluate object with lazy values
	if obj, ok := l.value.(map[string]any); ok {
		for key, value := range obj {
			lazy := NewLazyValue(value)

			forcedValue, err := lazy.Value()
			if err != nil {
				l.err = err
				return nil, l.err
			}

			obj[key] = forcedValue
		}
	}

	// Evaluate slice with lazy values
	if slice, ok := l.value.([]any); ok {
		for idx, value := range slice {
			lazy := NewLazyValue(value)

			forcedValue, err := lazy.Value()
			if err != nil {
				l.err = err
				return nil, l.err
			}

			slice[idx] = forcedValue
		}
	}

	return l.value, l.err
}

func (l *Lazy) Map(mapFn func(any) (any, error)) *Lazy {
	return NewLazy(func() (any, error) {
		value, err := l.Value()
		if err != nil {
			return nil, err
		}
		value, err = mapFn(value)
		if err != nil {
			return nil, err
		}
		return NewLazyValue(value).Value()
	})
}

func (l *Lazy) MapTo(value any) *Lazy {
	return l.Map(func(any) (any, error) {
		return NewLazyValue(value).Value()
	})
}
