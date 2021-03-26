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
		"event": { "type": "string", "enum" : ["pre_signup", "admin_api_create_user"] },
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
					"user.created.user_signup",
					"user.created.admin_api_create_user",
					"identity.created.user_add_identity",
					"identity.created.admin_api_add_identity",
					"identity.deleted.user_remove_identity",
					"identity.deleted.admin_api_remove_identity",
					"identity.updated.user_update_identity",
					"session.created.user_signup",
					"session.created.user_login",
					"session.created.user_promote_themselves",
					"session.deleted.user_revoke_session",
					"session.deleted.user_logout",
					"session.deleted.admin_api_revoke_session",
					"user.promoted.user_promote_themselves"
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
