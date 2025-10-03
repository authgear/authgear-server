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
		"event": { "type": "string", "enum" : [
			"user.pre_create",
			"user.profile.pre_update",
			"user.pre_schedule_deletion",
			"user.pre_schedule_anonymization",
			"oidc.jwt.pre_create",
			"oidc.id_token.pre_create",
			"authentication.pre_initialize",
			"authentication.post_identified",
			"authentication.pre_authenticated"

		] },
		"url": { "type": "string", "format": "x_hook_uri" }
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
					"user.reauthenticated",
					"user.profile.updated",
					"user.disabled",
					"user.reenabled",
					"user.anonymous.promoted",
					"user.deletion_scheduled",
					"user.deletion_unscheduled",
					"user.deleted",
					"user.anonymization_scheduled",
					"user.anonymization_unscheduled",
					"user.anonymized",
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
					"identity.oauth.disconnected",
					"identity.biometric.enabled",
					"identity.biometric.disabled"
				]
			}
		},
		"url": { "type": "string", "format": "x_hook_uri" }
	},
	"required": ["events", "url"]
}
`)

type NonBlockingHandlersConfig struct {
	Events []string `json:"events"`
	URL    string   `json:"url"`
}
