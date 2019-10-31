package template

type DefaultLoader struct {
	Map map[string]string
}

func NewDefaultLoader() *DefaultLoader {
	return &DefaultLoader{Map: make(map[string]string)}
}

func (s *DefaultLoader) Clone() *DefaultLoader {
	cloned := NewDefaultLoader()
	for key, value := range s.Map {
		cloned.Map[key] = value
	}
	return cloned
}

func (s *DefaultLoader) Load(name string) (string, error) {
	template, found := s.Map[name]
	if !found {
		return "", &errNotFound{name}
	}

	return template, nil
}
