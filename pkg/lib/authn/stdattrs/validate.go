package stdattrs

import (
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var Schema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"email": {
			"type": "string",
			"format": "email"
		},
		"phone_number": {
			"type": "string",
			"format": "phone"
		},
		"preferred_username": {
			"type": "string",
			"minLength": 1
		},
		"family_name": {
			"type": "string",
			"minLength": 1
		},
		"given_name": {
			"type": "string",
			"minLength": 1
		},
		"middle_name": {
			"type": "string",
			"minLength": 1
		},
		"name": {
			"type": "string",
			"minLength": 1
		},
		"nickname": {
			"type": "string",
			"minLength": 1
		},
		"picture": {
			"type": "string",
			"format": "uri"
		},
		"profile": {
			"type": "string",
			"format": "uri"
		},
		"website": {
			"type": "string",
			"format": "uri"
		},
		"gender": {
			"type": "string",
			"minLength": 1
		},
		"birthdate": {
			"type": "string",
			"format": "birthdate"
		},
		"zoneinfo": {
			"type": "string",
			"format": "timezone"
		},
		"locale": {
			"type": "string",
			"format": "bcp47"
		},
		"address": {
			"type": "object",
			"properties": {
				"formatted": {
					"type": "string",
					"minLength": 1
				},
				"street_address": {
					"type": "string",
					"minLength": 1
				},
				"locality": {
					"type": "string",
					"minLength": 1
				},
				"region": {
					"type": "string",
					"minLength": 1
				},
				"postal_code": {
					"type": "string",
					"minLength": 1
				},
				"country": {
					"type": "string",
					"minLength": 1
				}
			}
		}
	}
}
`)

func Validate(t T) error {
	a := t.ToClaims()
	return Schema.Validator().ValidateValue(a)
}
