package template

import (
	"fmt"
)

type Loader interface {
	Load(name string) (string, error)
}

type StringLoader struct {
	StringMap map[string]string
}

func NewStringLoader() *StringLoader {
	return &StringLoader{StringMap: make(map[string]string)}
}

func (s *StringLoader) Load(name string) (string, error) {
	template, found := s.StringMap[name]
	if !found {
		return "", fmt.Errorf("template with name `%s` not found", name)
	}

	return template, nil
}

type FSLoader struct {
	FilepathMap map[string]string
}

func NewFSLoader() *FSLoader {
	return &FSLoader{FilepathMap: make(map[string]string)}
}

func (f *FSLoader) Load(name string) (string, error) {
	filepath, found := f.FilepathMap[name]
	if !found {
		return "", fmt.Errorf("template with name `%s` not found", name)
	}

	return DownloadTemplateFromFilePath(filepath)
}

type HTTPLoader struct {
	URLMap map[string]string
}

func NewHTTPLoader() *HTTPLoader {
	return &HTTPLoader{URLMap: make(map[string]string)}
}

func (h *HTTPLoader) Load(name string) (string, error) {
	url, found := h.URLMap[name]
	if !found {
		return "", fmt.Errorf("template with name `%s` not found", name)
	}

	return DownloadTemplateFromURL(url)
}
