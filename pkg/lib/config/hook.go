package config

var _ = Schema.Add("HookConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sync_hook_timeout_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"sync_hook_total_timeout_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"blocking_handlers": { "type": "array", "items": { "$ref": "#/$defs/BlockingHookHandlersConfig" } },
		"non_blocking_handlers": { "type": "array", "items": { "$ref": "#/$defs/NonBlockingHookHandlersConfig" } }
	}
}
`)

type HookConfig struct {
	SyncTimeout         DurationSeconds             `json:"sync_hook_timeout_seconds,omitempty"`
	SyncTotalTimeout    DurationSeconds             `json:"sync_hook_total_timeout_seconds,omitempty"`
	BlockingHandlers    []BlockingHandlersConfig    `json:"blocking_handlers,omitempty"`
	NonBlockingHandlers []NonBlockingHandlersConfig `json:"non_blocking_handlers,omitempty"`
}

func (c *HookConfig) SetDefaults() {
	if c.SyncTimeout == 0 {
		c.SyncTimeout = DurationSeconds(5)
	}
	if c.SyncTotalTimeout == 0 {
		c.SyncTotalTimeout = DurationSeconds(10)
	}
}

var _ = Schema.Add("BlockingHookHandlersConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"event": { "type": "string", "enum" : ["user.pre_create"] },
		"url": { "type": "string", "format": "uri" }
	},
	"required": ["event", "url"]
}
`)

type BlockingHandlersConfig struct {
	Event string `json:"event"`
	URL   string `json:"url"`
}

var _ = Schema.Add("NonBlockingHookHandlersConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"events": {
			"type": "array",
			"items": {
				"type": "string",
				"enum" : [
					"*",
					"user.created",
					"user.authenticated",
					"user.anonymous.promoted",
					"identity.email.added",
					"identity.email.removed",
					"identity.email.updated",
					"identity.phone.added",
					"identity.phone.removed",
					"identity.phone.updated",
					"identity.username.added",
					"identity.username.removed",
					"identity.username.updated",
					"identity.oauth.connected",
					"identity.oauth.disconnected"
				]
			}
		},
		"url": { "type": "string", "format": "uri" }
	},
	"required": ["events", "url"]
}
`)

type NonBlockingHandlersConfig struct {
	Events []string `json:"events"`
	URL    string   `json:"url"`
}
