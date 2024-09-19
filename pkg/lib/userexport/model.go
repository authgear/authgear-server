package userexport

import (
	"encoding/json"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

// TODO: Set to 1000 as default after putting in worker task
const BatchSize = 3

type UserForExport struct {
	model.User

	Identities     []*identity.Info
	Authenticators []*authenticator.Info
}

type Error struct {
	Message string `json:"message,omitempty"`
	Reason  string `json:"reason,omitempty"`
}

var RequestSchema = validation.NewSimpleSchema(`
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"format": {
			"type": "string"
		}
	},
	"required": ["format"]
}
`)

type Request struct {
	Format string `json:"format"`
}

type Response struct {
	ID          string                `json:"id,omitempty"`
	CreatedAt   *time.Time            `json:"created_at,omitempty"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	FailedAt    *time.Time            `json:"failed_at,omitempty"`
	Status      redisqueue.TaskStatus `json:"status,omitempty"`
	Request     *Request              `json:"request,omitempty"`
	DownloadUrl string                `json:"download_url,omitempty"`
	Error       *Error                `json:"error,omitempty"`
}

type Address struct {
	Formatted     string `json:"formatted"`
	StreetAddress string `json:"street_address"`
	Locality      string `json:"locality"`
	Region        string `json:"region"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

type Identity struct {
	Type    model.IdentityType     `json:"type"`
	LoginID map[string]interface{} `json:"login_id,omitempty"`
	OAuth   map[string]interface{} `json:"oauth,omitempty"`
	LDAP    map[string]interface{} `json:"ldap,omitempty"`
	Claims  map[string]interface{} `json:"claims,omitempty"`
}

type MFATOTP struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

type MFA struct {
	Emails       []string   `json:"emails"`
	PhoneNumbers []string   `json:"phone_numbers"`
	TOTPs        []*MFATOTP `json:"totps"`
}

type Record struct {
	Sub string `json:"sub"`

	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	PhoneNumber       string `json:"phone_number"`

	EmailVerified       bool `json:"email_verified"`
	PhoneNumberVerified bool `json:"phone_number_verified"`

	Name       string   `json:"name"`
	GivenName  string   `json:"given_name"`
	FamilyName string   `json:"family_name"`
	MiddleName string   `json:"middle_name"`
	Nickname   string   `json:"nickname"`
	Profile    string   `json:"profile"`
	Picture    string   `json:"picture"`
	Website    string   `json:"website"`
	Gender     string   `json:"gender"`
	Birthdate  string   `json:"birthdate"`
	Zoneinfo   string   `json:"zoneinfo"`
	Locale     string   `json:"locale"`
	Address    *Address `json:"address"`

	CustomAttributes map[string]interface{} `json:"custom_attributes"`

	Roles  []string `json:"roles"`
	Groups []string `json:"groups"`

	Disabled bool `json:"disabled"`

	Identities []*Identity `json:"identities"`

	Mfa *MFA `json:"mfa"`

	BiometricCount int `json:"biometric_count"`
	PasskeyCount   int `json:"passkey_count"`
}

func NewResponseFromTask(task *redisqueue.Task) (*Response, error) {
	response := &Response{
		ID:          task.ID,
		CreatedAt:   task.CreatedAt,
		CompletedAt: task.CompletedAt,
		Status:      task.Status,
	}

	if task.Input != nil {
		var request Request
		err := json.Unmarshal(task.Input, &request)
		if err != nil {
			return nil, err
		}
		response.Request = &request
	}

	if task.Output != nil {
		var result Result
		err := json.Unmarshal(task.Output, &result)
		if err != nil {
			return nil, err
		}

		// TODO: sign a download url from filename
		response.DownloadUrl = result.Filename
	}

	return response, nil
}
