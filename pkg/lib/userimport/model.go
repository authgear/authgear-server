package userimport

import (
	"encoding/json"
	"fmt"

	"golang.org/x/exp/constraints"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var standardAttributeKeys []string = []string{
	"preferred_username",
	"email",
	"phone_number",
	"name",
	"given_name",
	"family_name",
	"middle_name",
	"nickname",
	"profile",
	"picture",
	"website",
	"gender",
	"birthdate",
	"zoneinfo",
	"locale",
	"address",
}

func mapGetNullable[M ~map[string]interface{}, T constraints.Ordered | ~bool](m M, key string) (*T, bool) {
	iface, ok := m[key]
	if !ok {
		return nil, false
	}
	if iface == nil {
		return nil, true
	}
	v, ok := iface.(T)
	if !ok {
		var t T
		panic(fmt.Errorf("%v is expected to be of type %T, but was %T", key, t, v))
	}
	return &v, true
}

func mapGetNonNull[M ~map[string]interface{}, T constraints.Ordered | ~bool](m M, key string) (T, bool) {
	var t T
	iface, ok := m[key]
	if !ok {
		return t, false
	}
	if iface == nil {
		panic(fmt.Errorf("%v is expected to be non-null", key))
	}
	v, ok := iface.(T)
	if !ok {
		panic(fmt.Errorf("%v is expected to be of type %T, but was %T", key, t, v))
	}
	return v, true
}

func mapGetNonNullMap[M ~map[string]interface{}, T ~map[string]interface{}](m M, key string) (T, bool) {
	iface, ok := m[key]
	if !ok {
		return nil, false
	}
	if iface == nil {
		panic(fmt.Errorf("%v is expected to be non-null", key))
	}
	v, ok := iface.(T)
	if !ok {
		var t T
		panic(fmt.Errorf("%v is expected to be of type %T, but was %T", key, t, v))
	}
	return v, true
}

func mapGetArrayOfNonNullItems[M ~map[string]interface{}, T constraints.Ordered | ~bool](m M, key string) ([]T, bool) {
	iface, ok := m[key]
	if !ok {
		return nil, false
	}

	var ts []T
	sliceIface, ok := iface.([]interface{})
	if !ok {
		panic(fmt.Errorf("%v is expected to be of type %T, but was %T", key, ts, iface))
	}

	for _, valueIface := range sliceIface {
		v, ok := valueIface.(T)
		if !ok {
			panic(fmt.Errorf("%v is expected to be of type %T, but was %T", key, ts, iface))
		}
		ts = append(ts, v)
	}

	return ts, true
}

const (
	IdentifierEmail             = "email"
	IdentifierPreferredUsername = "preferred_username"
	IdentifierPhoneNumber       = "phone_number"
)

const (
	PasswordTypeBcrypt = "bcrypt"
)

type Password map[string]interface{}

func (m Password) Type() string {
	return m["type"].(string)
}

func (m Password) PasswordHash() string {
	return m["password_hash"].(string)
}

type TOTP map[string]interface{}

func (m TOTP) Secret() string {
	return m["secret"].(string)
}

type MFA map[string]interface{}

func (m MFA) Email() (*string, bool) {
	return mapGetNullable[MFA, string](m, "email")
}

func (m MFA) PhoneNumber() (*string, bool) {
	return mapGetNullable[MFA, string](m, "phone_number")
}

func (m MFA) Password() (map[string]interface{}, bool) {
	return mapGetNonNullMap[MFA, map[string]interface{}](m, "password")
}

func (m MFA) TOTP() (map[string]interface{}, bool) {
	return mapGetNonNullMap[MFA, map[string]interface{}](m, "totp")
}

type Record map[string]interface{}

func (m Record) PreferredUsername() (*string, bool) {
	return mapGetNullable[Record, string](m, "preferred_username")
}

func (m Record) Email() (*string, bool) {
	return mapGetNullable[Record, string](m, "email")
}

func (m Record) PhoneNumber() (*string, bool) {
	return mapGetNullable[Record, string](m, "phone_number")
}

func (m Record) Disabled() (bool, bool) {
	return mapGetNonNull[Record, bool](m, "disabled")
}

func (m Record) EmailVerified() (bool, bool) {
	return mapGetNonNull[Record, bool](m, "email_verified")
}

func (m Record) PhoneNumberVerified() (bool, bool) {
	return mapGetNonNull[Record, bool](m, "phone_number_verified")
}

func (m Record) StandardAttributes() (map[string]interface{}, bool) {
	attrs := make(map[string]interface{})
	for key := range m {
		for _, k := range standardAttributeKeys {
			if key == k {
				attrs[key] = m[key]
			}
		}
	}
	if len(attrs) > 0 {
		return attrs, true
	}
	return nil, false
}

func (m Record) CustomAttributes() (map[string]interface{}, bool) {
	return mapGetNonNullMap[Record, map[string]interface{}](m, "custom_attributes")
}

func (m Record) Roles() ([]string, bool) {
	return mapGetArrayOfNonNullItems[Record, string](m, "roles")
}

func (m Record) Groups() ([]string, bool) {
	return mapGetArrayOfNonNullItems[Record, string](m, "groups")
}

func (m Record) Password() (map[string]interface{}, bool) {
	return mapGetNonNullMap[Record, map[string]interface{}](m, "password")
}

func (m Record) MFA() (map[string]interface{}, bool) {
	return mapGetNonNullMap[Record, map[string]interface{}](m, "mfa")
}

var RecordSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"preferred_username": { "type": ["string", "null"] },
		"email": { "type": ["string", "null"] },
		"phone_number": { "type": ["string", "null"] },

		"disabled": { "type": "boolean" },

		"email_verified": { "type": "boolean" },
		"phone_number_verified": { "type": "boolean" },

		"name": { "type": ["string", "null"] },
		"given_name": { "type": ["string", "null"] },
		"family_name": { "type": ["string", "null"] },
		"middle_name": { "type": ["string", "null"] },
		"nickname": { "type": ["string", "null"] },
		"profile": { "type": ["string", "null"] },
		"picture": { "type": ["string", "null"] },
		"website": { "type": ["string", "null"] },
		"gender": { "type": ["string", "null"] },
		"birthdate": { "type": ["string", "null"] },
		"zoneinfo": { "type": ["string", "null"] },
		"locale": { "type": ["string", "null"] },
		"address": { "type": ["object", "null"] },

		"custom_attributes": { "type": "object" },

		"roles": { "type": "array", "items": { "type": "string" } },
		"groups": { "type": "array", "items": { "type": "string" } },

		"password": {
			"type": "object",
			"properties": {
				"type": {
					"type": "string",
					"enum": ["bcrypt"]
				},
				"password_hash": {
					"type": "string"
				}
			},
			"required": ["type", "password_hash"]
		},

		"mfa": {
			"type": "object",
			"properties": {
				"email": { "type": ["string", "null"] },
				"phone_number": { "type": ["string", "null"] },
				"password": {
					"type": "object",
					"properties": {
						"type": {
							"type": "string",
							"enum": ["bcrypt"]
						},
						"password_hash": {
							"type": "string"
						}
					},
					"required": ["type", "password_hash"]
				},
				"totp": {
					"type": "object",
					"properties": {
						"secret": { "type": "string" }
					},
					"required": ["secret"]
				}
			}
		}
	}
}
`)

type Request struct {
	Upsert     bool   `json:"upsert,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	// Records is json.RawMessage because we want to delay the deserialization until we actually process the record.
	Records []json.RawMessage `json:"records,omitempty"`
}

var RequestSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"upsert": {
			"type": "boolean"
		},
		"identifier": {
			"type": "string",
			"enum": ["preferred_username", "email", "phone_number"]
		},
		"records": {
			"type": "array",
			"minItems": 1,
			"items": {
				"type": "object"
			}
		}
	},
	"required": ["identifier", "records"]
}
`)

type Options struct {
	Upsert     bool
	Identifier string
}

type Warning struct {
	Message string `json:"message,omitempty"`
}

type Outcome string

const (
	OutcomeInserted Outcome = "inserted"
	OutcomeUpdated  Outcome = "updated"
	OutcomeSkipped  Outcome = "skipped"
	OutcomeFailed   Outcome = "failed"
)

type Summary struct {
	Total    int `json:"total"`
	Inserted int `json:"inserted"`
	Updated  int `json:"updated"`
	Skipped  int `json:"skipped"`
	Failed   int `json:"failed"`
}

type Detail struct {
	Index    int                   `json:"index"`
	Record   json.RawMessage       `json:"record"`
	Warnings []Warning             `json:"warnings,omitempty"`
	Errors   []*apierrors.APIError `json:"errors,omitempty"`
}
