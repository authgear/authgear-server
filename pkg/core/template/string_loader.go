package template

type StringLoader struct {
	StringMap map[string]string
}

func NewStringLoader() *StringLoader {
	return &StringLoader{StringMap: make(map[string]string)}
}

func (s *StringLoader) Load(name string) (string, error) {
	template, found := s.StringMap[name]
	if !found {
		return "", &errNotFound{name}
	}

	return template, nil
}
