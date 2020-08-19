package utils

type LoadFunc func(keys []interface{}) ([]interface{}, error)

type promise struct {
	key     interface{}
	settled bool

	value interface{}
	err   error
}

type DataLoader struct {
	MaxBatch int
	loadFn   LoadFunc
	cache    map[interface{}]*promise
	queue    []*promise
}

func NewDataLoader(loadFn LoadFunc) *DataLoader {
	return &DataLoader{
		MaxBatch: 20,
		loadFn:   loadFn,
		cache:    make(map[interface{}]*promise),
	}
}

func (l *DataLoader) run() {
	keys := make([]interface{}, len(l.queue))
	for i, p := range l.queue {
		keys[i] = p.key
	}
	values, err := l.loadFn(keys)
	for i, p := range l.queue {
		p.value = values[i]
		p.err = err
		p.settled = true
	}
	l.queue = nil
}

func (l *DataLoader) Load(key interface{}) func() (interface{}, error) {
	p, ok := l.cache[key]
	if !ok {
		if len(l.queue) >= l.MaxBatch {
			l.run()
		}

		p = &promise{key: key}
		l.queue = append(l.queue, p)
		l.cache[p.key] = p
	}
	return func() (interface{}, error) {
		if !p.settled {
			l.run()
		}
		return p.value, p.err
	}
}
