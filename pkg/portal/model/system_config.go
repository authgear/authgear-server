package model

type SystemConfig struct {
	AuthgearClientID        string      `json:"authgearClientID"`
	AuthgearEndpoint        string      `json:"authgearEndpoint"`
	AppHostSuffix           string      `json:"appHostSuffix"`
	PossibleTemplateLocales []string    `json:"possibleTemplateLocales"`
	Themes                  interface{} `json:"themes,omitempty"`
	Translations            interface{} `json:"translations,omitempty"`
}
