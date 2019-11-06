package template

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type DefaultLoader struct {
	Map map[config.TemplateItemType]T
}

func NewDefaultLoader() *DefaultLoader {
	return &DefaultLoader{Map: make(map[config.TemplateItemType]T)}
}

func (s *DefaultLoader) Clone() *DefaultLoader {
	cloned := NewDefaultLoader()
	for key, value := range s.Map {
		cloned.Map[key] = value
	}
	return cloned
}

func (s *DefaultLoader) Load(templateType config.TemplateItemType) (string, error) {
	template, found := s.Map[templateType]
	if !found {
		return "", &errNotFound{string(templateType)}
	}

	return template.Default, nil
}
