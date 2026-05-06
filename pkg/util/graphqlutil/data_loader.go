package graphqlutil

import (
	"context"
)

// LoadFunc must satisfy the following conditions.
//
// 1. The length of the result must match that of keys.
// 2. If a given key resolves to nothing, nil must be returned instead.
// 3. The order of the result must match that of keys.
//
// So it is the responsibility of LoadFunc to satisfy these conditions.
// The underlying implementation used by LoadFunc may not satisfy the conditions.
type LoadFunc func(ctx context.Context, keys []any) ([]any, error)

type dataLoaderTask struct {
	key    any
	settle func(value any, err error)
}

type DataLoaderInterface interface {
	Load(ctx context.Context, key any) *Lazy
	LoadMany(ctx context.Context, keys []any) *Lazy
	Clear(key any)
	ClearAll()
	Prime(key any, value any)
}

type DataLoader struct {
	MaxBatch int
	loadFn   LoadFunc
	cache    map[any]*Lazy
	queue    []dataLoaderTask
}

func NewDataLoader(loadFn LoadFunc) *DataLoader {
	return &DataLoader{
		MaxBatch: 20,
		loadFn:   loadFn,
		cache:    make(map[any]*Lazy),
	}
}

func (l *DataLoader) run(ctx context.Context) {
	keys := make([]any, len(l.queue))
	for i, p := range l.queue {
		keys[i] = p.key
	}
	values, err := l.loadFn(ctx, keys)
	for i, p := range l.queue {
		if err != nil {
			p.settle(nil, err)
		} else {
			p.settle(values[i], nil)
		}
	}
	l.queue = nil
}

func (l *DataLoader) Load(ctx context.Context, key any) *Lazy {
	p, ok := l.cache[key]
	if !ok {
		if len(l.queue) >= l.MaxBatch {
			l.run(ctx)
		}

		settled := false
		var value any
		var err error
		p = NewLazy(func() (any, error) {
			if !settled {
				l.run(ctx)
			}
			return value, err
		})
		l.queue = append(l.queue, dataLoaderTask{
			key: key,
			settle: func(v any, e error) {
				value = v
				err = e
				settled = true
			},
		})
		l.cache[key] = p
	}
	return p
}

func (l *DataLoader) LoadMany(ctx context.Context, keys []any) *Lazy {
	values := make([]any, len(keys))
	for idx, key := range keys {
		value := l.Load(ctx, key)
		values[idx] = value
	}
	return NewLazyValue(values)
}

func (l *DataLoader) Clear(key any) {
	delete(l.cache, key)
}

func (l *DataLoader) ClearAll() {
	l.cache = make(map[any]*Lazy)
}

func (l *DataLoader) Prime(key any, value any) {
	_, ok := l.cache[key]
	if ok {
		return
	}
	l.cache[key] = NewLazyValue(value)
}
