package config

var _ = Schema.Add("UserProfileJSONPointer", `
{
	"type": "object",
	"required": ["user_profile"],
	"additionalProperties": false,
	"properties": {
		"user_profile": {
			"type": "object",
			"required": ["pointer"],
			"additionalProperties": false,
			"properties": {
				"pointer": { "type": "string", "format": "json-pointer" }
			}
		}
	}
}
`)

type UserProfileJSONPointer struct {
	UserProfile *JSONPointer `json:"user_profile,omitempty"`
}

type JSONPointer struct {
	Pointer string `json:"pointer,omitempty"`
}
