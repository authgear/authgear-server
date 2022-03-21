package config

var _ = Schema.Add("GoogleTagManagerConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"container_id": {
			"type": "string",
			"format": "google_tag_manager_container_id"
		}
	}
}
`)

type GoogleTagManagerConfig struct {
	ContainerID string `json:"container_id,omitempty"`
}
