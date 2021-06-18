package model

type SystemConfig struct {
	AuthgearClientID   string      `json:"authgearClientID"`
	AuthgearEndpoint   string      `json:"authgearEndpoint"`
	AppHostSuffix      string      `json:"appHostSuffix"`
	AvailableLanguages []string    `json:"availableLanguages"`
	BuiltinLanguages   []string    `json:"builtinLanguages"`
	Themes             interface{} `json:"themes,omitempty"`
	Translations       interface{} `json:"translations,omitempty"`
	SearchEnabled      bool        `json:"searchEnabled"`
	AuditLogEnabled    bool        `json:"auditLogEnabled"`
}
