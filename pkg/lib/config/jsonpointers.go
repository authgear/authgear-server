package config

import "github.com/iawaknahc/jsonschema/pkg/jsonpointer"

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
	UserProfile *JSONPointer `json:"user_profile,omitempty" nullable:"true"`
}

type JSONPointer struct {
	Pointer string `json:"pointer,omitempty"`
}

func (p JSONPointer) MustGetJSONPointer() jsonpointer.T {
	pointer := jsonpointer.MustParse(string(p.Pointer))
	return pointer
}
