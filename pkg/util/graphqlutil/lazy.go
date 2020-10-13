package graphqlutil

type Lazy struct {
	init  func() (interface{}, error)
	value interface{}
	err   error
}

func NewLazy(init func() (interface{}, error)) *Lazy {
	return &Lazy{init: init}
}

func NewLazyValue(value interface{}) *Lazy {
	switch value := value.(type) {
	case *Lazy:
		return value
	case func() (interface{}, error):
		return NewLazy(value)
	default:
		return &Lazy{value: value}
	}
}

func (l *Lazy) Value() (interface{}, error) {
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
	if obj, ok := l.value.(map[string]interface{}); ok {
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
	if slice, ok := l.value.([]interface{}); ok {
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

func (l *Lazy) Map(mapFn func(interface{}) (interface{}, error)) *Lazy {
	return NewLazy(func() (interface{}, error) {
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

func (l *Lazy) MapTo(value interface{}) *Lazy {
	return l.Map(func(interface{}) (interface{}, error) {
		return NewLazyValue(value).Value()
	})
}
