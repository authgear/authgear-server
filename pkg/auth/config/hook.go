package config

var _ = Schema.Add("HookConfig", `
{
	"type": "object",
	"properties": {
		"sync_hook_timeout_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"sync_hook_total_timeout_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"handlers": { "type": "array", "items": { "$ref": "HookHandlerConfig" } }
	}
}
`)

type HookConfig struct {
	SyncTimeout      DurationSeconds     `json:"sync_hook_timeout_seconds,omitempty"`
	SyncTotalTimeout DurationSeconds     `json:"sync_hook_total_timeout_seconds,omitempty"`
	Handlers         []HookHandlerConfig `json:"handlers,omitempty"`
}

var _ = Schema.Add("HookHandlerConfig", `
{
	"type": "object",
	"properties": {
		"event": { "type": "string" },
		"url": { "type": "string", "format": "uri" }
	}
}
`)

type HookHandlerConfig struct {
	Event string `json:"event"`
	URL   string `json:"url"`
}
