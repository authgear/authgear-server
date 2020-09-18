package graphqlutil

// LoadFunc must satisfy the following conditions.
//
// 1. The length of the result must match that of keys.
// 2. If a given key resolves to nothing, nil must be returned instead.
// 3. The order of the result must match that of keys.
//
// So it is the responsibility of LoadFunc to satisfy these conditions.
// The underlying implementation used by LoadFunc may not satisfy the conditions.
type LoadFunc func(keys []interface{}) ([]interface{}, error)

type dataLoaderTask struct {
	key    interface{}
	settle func(value interface{}, err error)
}

type DataLoader struct {
	MaxBatch int
	loadFn   LoadFunc
	cache    map[interface{}]*Lazy
	queue    []dataLoaderTask
}

func NewDataLoader(loadFn LoadFunc) *DataLoader {
	return &DataLoader{
		MaxBatch: 20,
		loadFn:   loadFn,
		cache:    make(map[interface{}]*Lazy),
	}
}

func (l *DataLoader) run() {
	keys := make([]interface{}, len(l.queue))
	for i, p := range l.queue {
		keys[i] = p.key
	}
	values, err := l.loadFn(keys)
	for i, p := range l.queue {
		if err != nil {
			p.settle(nil, err)
		} else {
			p.settle(values[i], nil)
		}
	}
	l.queue = nil
}

func (l *DataLoader) Reset(key interface{}) {
	if l == nil {
		return
	}
	delete(l.cache, key)
}

func (l *DataLoader) Load(key interface{}) *Lazy {
	p, ok := l.cache[key]
	if !ok {
		if len(l.queue) >= l.MaxBatch {
			l.run()
		}

		settled := false
		var value interface{}
		var err error
		p = NewLazy(func() (interface{}, error) {
			if !settled {
				l.run()
			}
			return value, err
		})
		l.queue = append(l.queue, dataLoaderTask{
			key: key,
			settle: func(v interface{}, e error) {
				value = v
				err = e
				settled = true
			},
		})
		l.cache[key] = p
	}
	return p
}
