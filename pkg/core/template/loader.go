package template

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
		return "", &errNotFound{name}
	}

	return template, nil
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
		return "", &errNotFound{name}
	}

	return DownloadTemplateFromURL(url)
}
