package config

var _ = Schema.Add("HookConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sync_hook_timeout_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"sync_hook_total_timeout_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"handlers": { "type": "array", "items": { "$ref": "#/$defs/HookHandlerConfig" } }
	}
}
`)

type HookConfig struct {
	SyncTimeout      DurationSeconds     `json:"sync_hook_timeout_seconds,omitempty"`
	SyncTotalTimeout DurationSeconds     `json:"sync_hook_total_timeout_seconds,omitempty"`
	Handlers         []HookHandlerConfig `json:"handlers,omitempty"`
}

func (c *HookConfig) SetDefaults() {
	if c.SyncTimeout == 0 {
		c.SyncTimeout = DurationSeconds(5)
	}
	if c.SyncTotalTimeout == 0 {
		c.SyncTotalTimeout = DurationSeconds(10)
	}
}

var _ = Schema.Add("HookHandlerConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"event": { "type": "string" },
		"url": { "type": "string", "format": "uri" }
	},
	"required": ["event", "url"]
}
`)

type HookHandlerConfig struct {
	Event string `json:"event"`
	URL   string `json:"url"`
}
