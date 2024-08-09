package userimport

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	"golang.org/x/exp/constraints"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

// BodyMaxSize is 500KB.
var BodyMaxSize int64 = 500 * 1000

const RedactPlaceholder = "REDACTED"

var RecordSchemaForIdentifierEmail *validation.SimpleSchema
var RecordSchemaForIdentifierPhoneNumber *validation.SimpleSchema
var RecordSchemaForIdentifierPreferredUsername *validation.SimpleSchema

func init() {
	nullString := validation.SchemaBuilder{}.
		Types(validation.TypeString, validation.TypeNull)
	str := validation.SchemaBuilder{}.
		Type(validation.TypeString)

	makeBase := func() validation.SchemaBuilder {
		boolean := validation.SchemaBuilder{}.
			Type(validation.TypeBoolean)

		nullObject := validation.SchemaBuilder{}.
			Types(validation.TypeObject, validation.TypeNull)

		customAttributes := validation.SchemaBuilder{}.
			Type(validation.TypeObject)

		rolesOrGroups := validation.SchemaBuilder{}.
			Type(validation.TypeArray).
			Items(str)

		password := validation.SchemaBuilder{}.
			Type(validation.TypeObject).
			AdditionalPropertiesFalse().
			Required("type", "password_hash")
		password.Properties().
			Property("type", validation.SchemaBuilder{}.Type(validation.TypeString).Enum("bcrypt")).
			Property("password_hash", str).
			Property("expire_after", validation.SchemaBuilder{}.Type(validation.TypeString).Format("date-time"))

		totp := validation.SchemaBuilder{}.
			Type(validation.TypeObject).
			AdditionalPropertiesFalse().
			Required("secret")
		totp.Properties().
			Property("secret", str)

		mfa := validation.SchemaBuilder{}.
			Type(validation.TypeObject).
			AdditionalPropertiesFalse()
		mfa.Properties().
			Property("email", nullString).
			Property("phone_number", nullString).
			Property("password", password).
			Property("totp", totp)

		baseSchema := validation.SchemaBuilder{}.
			Type(validation.TypeObject).
			AdditionalPropertiesFalse()

		baseSchema.Properties().
			Property("disabled", boolean).
			Property("email_verified", boolean).
			Property("phone_number_verified", boolean).
			Property("name", nullString).
			Property("given_name", nullString).
			Property("family_name", nullString).
			Property("middle_name", nullString).
			Property("nickname", nullString).
			Property("profile", nullString).
			Property("picture", nullString).
			Property("website", nullString).
			Property("gender", nullString).
			Property("birthdate", nullString).
			Property("zoneinfo", nullString).
			Property("locale", nullString).
			Property("address", nullObject).
			Property("custom_attributes", customAttributes).
			Property("roles", rolesOrGroups).
			Property("groups", rolesOrGroups).
			Property("password", password).
			Property("mfa", mfa)

		return baseSchema
	}

	email := makeBase().
		Required("email")
	email.Properties().
		Property("email", str).
		Property("phone_number", nullString).
		Property("preferred_username", nullString)
	RecordSchemaForIdentifierEmail = email.ToSimpleSchema()

	phoneNumber := makeBase().
		Required("phone_number")
	phoneNumber.Properties().
		Property("phone_number", str).
		Property("email", nullString).
		Property("preferred_username", nullString)
	RecordSchemaForIdentifierPhoneNumber = phoneNumber.ToSimpleSchema()

	preferredUsername := makeBase().
		Required("preferred_username")
	preferredUsername.Properties().
		Property("preferred_username", str).
		Property("email", nullString).
		Property("phone_number", nullString)
	RecordSchemaForIdentifierPreferredUsername = preferredUsername.ToSimpleSchema()
}

var standardAttributeKeys []string = []string{
	// Note we don't need IdentityAware stdAttr ["email", "phone", "preferred_username"] here, since they are already populated in StdAttrsService
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

func (m Password) ExpireAfter() *time.Time {
	var dt interface{}
	var ok bool
	if dt, ok = m["expire_after"]; !ok {
		return nil
	}
	t, err := time.Parse(time.RFC3339, dt.(string))
	if err != nil {
		// The json schema validation should already ensure it is in correct format.
		// If it is not valid, it should be a panic.
		panic(err)
	}
	return &t
}

func (m Password) Redact() {
	m["password_hash"] = RedactPlaceholder
}

type TOTP map[string]interface{}

func (m TOTP) Redact() {
	m["secret"] = RedactPlaceholder
}

func (m TOTP) Secret() string {
	return m["secret"].(string)
}

type MFA map[string]interface{}

func (m MFA) Redact() {
	if password, ok := m.Password(); ok {
		Password(password).Redact()
	}
	if totp, ok := m.TOTP(); ok {
		TOTP(totp).Redact()
	}
}

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

func (m Record) Redact() {
	if password, ok := m.Password(); ok {
		Password(password).Redact()
	}
	if mfa, ok := m.MFA(); ok {
		MFA(mfa).Redact()
	}
}

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

func (m Record) standardAttributes() (map[string]interface{}, bool) {
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

func (m Record) StandardAttributesList() (attrsList attrs.List) {
	stdAttrs, ok := m.standardAttributes()
	if !ok {
		return
	}

	for key, value := range stdAttrs {
		ptr := jsonpointer.T{key}.String()
		attrsList = append(attrsList, attrs.T{
			Pointer: ptr,
			Value:   value,
		})
	}
	return
}

func (m Record) customAttributes() (map[string]interface{}, bool) {
	return mapGetNonNullMap[Record, map[string]interface{}](m, "custom_attributes")
}

func (m Record) CustomAttributesList() (attrsList attrs.List) {
	customAttrs, ok := m.customAttributes()
	if !ok {
		return
	}

	for key, value := range customAttrs {
		ptr := jsonpointer.T{key}.String()
		attrsList = append(attrsList, attrs.T{
			Pointer: ptr,
			Value:   value,
		})
	}
	return
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

func (o *Options) RecordSchema() *validation.SimpleSchema {
	switch o.Identifier {
	case IdentifierEmail:
		return RecordSchemaForIdentifierEmail
	case IdentifierPhoneNumber:
		return RecordSchemaForIdentifierPhoneNumber
	case IdentifierPreferredUsername:
		return RecordSchemaForIdentifierPreferredUsername
	default:
		panic(fmt.Errorf("unknown identifier: %v", o.Identifier))
	}
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
	Record   Record                `json:"record,omitempty"`
	Outcome  Outcome               `json:"outcome,omitempty"`
	UserID   string                `json:"user_id,omitempty"`
	Warnings []Warning             `json:"warnings,omitempty"`
	Errors   []*apierrors.APIError `json:"errors,omitempty"`
}

type Result struct {
	Summary *Summary `json:"summary,omitempty"`
	Details []Detail `json:"details,omitempty"`
}

type Response struct {
	ID          string                `json:"id,omitempty"`
	CreatedAt   *time.Time            `json:"created_at,omitempty"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	Status      redisqueue.TaskStatus `json:"status,omitempty"`
	Summary     *Summary              `json:"summary,omitempty"`
	Details     []Detail              `json:"details,omitempty"`
}

func NewResponseFromTask(task *redisqueue.Task) (*Response, error) {
	response := &Response{
		ID:          task.ID,
		CreatedAt:   task.CreatedAt,
		CompletedAt: task.CompletedAt,
		Status:      task.Status,
	}

	if task.Output != nil {
		var result Result
		err := json.Unmarshal(task.Output, &result)
		if err != nil {
			return nil, err
		}

		response.Summary = result.Summary
		response.Details = result.Details
	}

	return response, nil
}
