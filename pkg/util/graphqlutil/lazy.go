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

		if lazy, ok := value.(*Lazy); ok {
			return lazy.Value()
		}
		return value, nil
	})
}
