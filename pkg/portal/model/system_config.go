package model

type SystemConfig struct {
	AuthgearClientID         string      `json:"authgearClientID"`
	AuthgearEndpoint         string      `json:"authgearEndpoint"`
	AppHostSuffix            string      `json:"appHostSuffix"`
	SupportedResourceLocales []string    `json:"supportedResourceLocales"`
	Themes                   interface{} `json:"themes,omitempty"`
	Translations             interface{} `json:"translations,omitempty"`
}
