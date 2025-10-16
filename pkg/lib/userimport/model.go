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
	reusedSchemaBuilders := reuseSchemaBuilders()
	str := validation.SchemaBuilder{}.
		Type(validation.TypeString)

	minLenStr := str.MinLength(1)

	makeBase := func() validation.SchemaBuilder {
		boolean := validation.SchemaBuilder{}.
			Type(validation.TypeBoolean)

		rfc3339 := validation.SchemaBuilder{}.
			Type(validation.TypeString).
			Format("date-time")

		customAttributes := validation.SchemaBuilder{}.
			Type(validation.TypeObject)

		rolesOrGroups := validation.SchemaBuilder{}.
			Type(validation.TypeArray).
			Items(minLenStr)

		password := validation.SchemaBuilder{}.
			Type(validation.TypeObject).
			AdditionalPropertiesFalse().
			Required("type", "password_hash")
		password.Properties().
			Property("type", validation.SchemaBuilder{}.Type(validation.TypeString).Enum("bcrypt")).
			Property("password_hash", minLenStr).
			Property("expire_after", rfc3339)

		totp := validation.SchemaBuilder{}.
			Type(validation.TypeObject).
			AdditionalPropertiesFalse().
			Required("secret")
		totp.Properties().
			Property("secret", minLenStr)

		mfa := validation.SchemaBuilder{}.
			Type(validation.TypeObject).
			AdditionalPropertiesFalse()
		mfa.Properties().
			Property("email", reusedSchemaBuilders.Email.AddTypeNull()).
			Property("phone_number", reusedSchemaBuilders.PhoneNumber.AddTypeNull()).
			Property("password", password).
			Property("totp", totp)

		baseSchema := validation.SchemaBuilder{}.
			Type(validation.TypeObject).
			AdditionalPropertiesFalse()

		baseSchema.Properties().
			Property("disabled", boolean).
			Property("account_valid_from", rfc3339).
			Property("account_valid_until", rfc3339).
			Property("email_verified", boolean).
			Property("phone_number_verified", boolean).
			Property("name", reusedSchemaBuilders.Name.AddTypeNull()).
			Property("given_name", reusedSchemaBuilders.GivenName.AddTypeNull()).
			Property("family_name", reusedSchemaBuilders.FamilyName.AddTypeNull()).
			Property("middle_name", reusedSchemaBuilders.MiddleName.AddTypeNull()).
			Property("nickname", reusedSchemaBuilders.Nickname.AddTypeNull()).
			Property("profile", reusedSchemaBuilders.Profile.AddTypeNull()).
			Property("picture", reusedSchemaBuilders.Picture.AddTypeNull()).
			Property("website", reusedSchemaBuilders.Website.AddTypeNull()).
			Property("gender", reusedSchemaBuilders.Gender.AddTypeNull()).
			Property("birthdate", reusedSchemaBuilders.Birthdate.AddTypeNull()).
			Property("zoneinfo", reusedSchemaBuilders.Zoneinfo.AddTypeNull()).
			Property("locale", reusedSchemaBuilders.Locale.AddTypeNull()).
			Property("address", reusedSchemaBuilders.Address.AddTypeNull()).
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
		Property("email", reusedSchemaBuilders.Email).
		Property("phone_number", reusedSchemaBuilders.PhoneNumber.AddTypeNull()).
		Property("preferred_username", reusedSchemaBuilders.PreferredUsername.AddTypeNull())
	RecordSchemaForIdentifierEmail = email.ToSimpleSchema()

	phoneNumber := makeBase().
		Required("phone_number")
	phoneNumber.Properties().
		Property("phone_number", reusedSchemaBuilders.PhoneNumber).
		Property("email", reusedSchemaBuilders.Email.AddTypeNull()).
		Property("preferred_username", reusedSchemaBuilders.PreferredUsername.AddTypeNull())
	RecordSchemaForIdentifierPhoneNumber = phoneNumber.ToSimpleSchema()

	preferredUsername := makeBase().
		Required("preferred_username")
	preferredUsername.Properties().
		Property("preferred_username", reusedSchemaBuilders.PreferredUsername).
		Property("email", reusedSchemaBuilders.Email.AddTypeNull()).
		Property("phone_number", reusedSchemaBuilders.PhoneNumber.AddTypeNull())
	RecordSchemaForIdentifierPreferredUsername = preferredUsername.ToSimpleSchema()
}

var nonIdentityAwareStandardAttributeKeys []string = []string{
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

func mapGetNullableMap[M ~map[string]interface{}, T ~map[string]interface{}](m M, key string) (value T, exist bool, ok bool) {
	iface, ok := m[key]
	if !ok {
		return nil, false, false
	}
	if iface == nil {
		return nil, false, false
	}
	v, ok := iface.(T)
	if !ok {
		return nil, true, false
	}
	return v, true, true
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

func mapGetRFC3339InUTC[M ~map[string]interface{}](m M, key string) (*time.Time, bool) {
	var iface interface{}
	iface, ok := m[key]
	if !ok {
		return nil, false
	}
	if iface == nil {
		panic(fmt.Errorf("%v is expected to be non-null", key))
	}

	str, ok := iface.(string)
	if !ok {
		panic(fmt.Errorf("%v is expected to be a RFC3339 timestamp, but was %T", key, iface))
	}

	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		// The json schema validation should already ensure it is in correct format.
		// If it is not valid, it should be a panic.
		panic(err)
	}

	t = t.In(time.UTC)
	return &t, true
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
	t, _ := mapGetRFC3339InUTC(m, "expire_after")
	return t
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
	if password, exist, ok := m.MaybePassword(); ok {
		Password(password).Redact()
	} else if exist {
		m["password"] = RedactPlaceholder
	}
	if totp, exist, ok := m.MaybeTOTP(); ok {
		TOTP(totp).Redact()
	} else if exist {
		m["totp"] = RedactPlaceholder
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

func (m MFA) MaybePassword() (value map[string]interface{}, exist bool, ok bool) {
	return mapGetNullableMap[MFA, map[string]interface{}](m, "password")
}

func (m MFA) TOTP() (map[string]interface{}, bool) {
	return mapGetNonNullMap[MFA, map[string]interface{}](m, "totp")
}

func (m MFA) MaybeTOTP() (value map[string]interface{}, exist bool, ok bool) {
	return mapGetNullableMap[MFA, map[string]interface{}](m, "totp")
}

type Record map[string]interface{}

func (m Record) Redact() {
	if password, exist, ok := m.MaybePassword(); ok {
		Password(password).Redact()
	} else if exist {
		m["password"] = RedactPlaceholder
	}
	if mfa, exist, ok := m.MaybeMFA(); ok {
		MFA(mfa).Redact()
	} else if exist {
		m["mfa"] = RedactPlaceholder
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

func (m Record) AccountValidFrom() (*time.Time, bool) {
	return mapGetRFC3339InUTC(m, "account_valid_from")
}

func (m Record) AccountValidUntil() (*time.Time, bool) {
	return mapGetRFC3339InUTC(m, "account_valid_until")
}

func (m Record) EmailVerified() (bool, bool) {
	return mapGetNonNull[Record, bool](m, "email_verified")
}

func (m Record) PhoneNumberVerified() (bool, bool) {
	return mapGetNonNull[Record, bool](m, "phone_number_verified")
}

func (m Record) nonIdentityAwareStandardAttributes() (map[string]interface{}, bool) {
	attrs := make(map[string]interface{})
	for key := range m {
		for _, k := range nonIdentityAwareStandardAttributeKeys {
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

func (m Record) NonIdentityAwareStandardAttributesList() (attrsList attrs.List) {
	stdAttrs, ok := m.nonIdentityAwareStandardAttributes()
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

func (m Record) MaybePassword() (value map[string]interface{}, exist bool, ok bool) {
	return mapGetNullableMap[Record, map[string]interface{}](m, "password")
}

func (m Record) MFA() (map[string]interface{}, bool) {
	return mapGetNonNullMap[Record, map[string]interface{}](m, "mfa")
}

func (m Record) MaybeMFA() (value map[string]interface{}, exist bool, ok bool) {
	return mapGetNullableMap[Record, map[string]interface{}](m, "mfa")
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
	Error       *apierrors.APIError   `json:"error,omitempty"`
	Summary     *Summary              `json:"summary,omitempty"`
	Details     []Detail              `json:"details,omitempty"`
}

func NewResponseFromJob(job *Job) *Response {
	return &Response{
		ID:          job.ID,
		CreatedAt:   &job.CreatedAt,
		CompletedAt: &job.CreatedAt,
		Status:      redisqueue.TaskStatusCompleted,
		Summary:     &Summary{},
	}
}

func NewResponseFromTask(task *redisqueue.Task) (*Response, error) {
	response := &Response{
		ID:          task.ID,
		CreatedAt:   task.CreatedAt,
		CompletedAt: task.CompletedAt,
		Status:      task.Status,
		Error:       task.Error,
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

func (r *Response) AggregateTaskResult(taskOffset int, task *redisqueue.Task) error {
	if task.CompletedAt == nil {
		// Any task not yet complete => Job not yet complete
		r.CompletedAt = nil
		r.Status = redisqueue.TaskStatusPending
		// No summary if job is still pending
		r.Summary = nil

		// Clear details to avoid out-of-ordered records.
		r.Details = nil
	} else if r.CompletedAt != nil && task.CompletedAt.After(*r.CompletedAt) {
		// Use the latest completion timestamp of tasks.
		r.CompletedAt = task.CompletedAt
	}

	if task.Error != nil {
		// TODO: merge errors from tasks?
		r.Error = task.Error
	}

	if task.Output != nil {
		var result Result
		err := json.Unmarshal(task.Output, &result)
		if err != nil {
			return err
		}

		if r.Summary != nil {
			r.Summary.Total += result.Summary.Total
			r.Summary.Inserted += result.Summary.Inserted
			r.Summary.Updated += result.Summary.Updated
			r.Summary.Skipped += result.Summary.Skipped
			r.Summary.Failed += result.Summary.Failed
		}
		if r.CompletedAt != nil {
			for _, detail := range result.Details {
				adjustedDetail := detail
				adjustedDetail.Index = adjustedDetail.Index + taskOffset

				r.Details = append(r.Details, adjustedDetail)
			}
		}
	}

	return nil
}

type Job struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	TaskIDs   []string  `json:"task_ids"`
}
