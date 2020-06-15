package config

type TemplateConfig struct {
	Items []TemplateItem `json:"items"`
}

type TemplateItemType string

type TemplateItem struct {
	Type        TemplateItemType `json:"type,omitempty"`
	LanguageTag string           `json:"language_tag,omitempty"`
	Key         string           `json:"key,omitempty"`
	URI         string           `json:"uri,omitempty"`
}
