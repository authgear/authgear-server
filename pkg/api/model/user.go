package model

import (
	"time"
)

type User struct {
	Meta
	LastLoginAt        *time.Time             `json:"last_login_at,omitempty"`
	IsAnonymous        bool                   `json:"is_anonymous"`
	IsVerified         bool                   `json:"is_verified"`
	IsDisabled         bool                   `json:"is_disabled"`
	DisableReason      *string                `json:"disable_reason,omitempty"`
	IsDeactivated      bool                   `json:"is_deactivated"`
	DeleteAt           *time.Time             `json:"delete_at,omitempty"`
	IsAnonymized       bool                   `json:"is_anonymized"`
	AnonymizeAt        *time.Time             `json:"anonymize_at,omitempty"`
	CanReauthenticate  bool                   `json:"can_reauthenticate"`
	StandardAttributes map[string]interface{} `json:"standard_attributes,omitempty"`
	CustomAttributes   map[string]interface{} `json:"custom_attributes,omitempty"`
	Web3               *UserWeb3Info          `json:"x_web3,omitempty"`
	Roles              []string               `json:"roles,omitempty"`
	Groups             []string               `json:"groups,omitempty"`
	MFAEnrollmentEndAt *time.Time             `json:"mfa_enrollment_end_at,omitempty"`
}

func (u *User) EndUserAccountID() string {
	if s, ok := u.StandardAttributes[string(ClaimEmail)].(string); ok && s != "" {
		return s
	}
	if s, ok := u.StandardAttributes[string(ClaimPreferredUsername)].(string); ok && s != "" {
		return s
	}
	if s, ok := u.StandardAttributes[string(ClaimPhoneNumber)].(string); ok && s != "" {
		return s
	}
	if u.Web3 != nil && len(u.Web3.Accounts) > 0 {
		first := u.Web3.Accounts[0]
		return first.EndUserAccountID()
	}

	return ""
}

type UserRef struct {
	Meta
}

type ElasticsearchUserRaw struct {
	ID                 string
	AppID              string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	LastLoginAt        *time.Time
	IsDisabled         bool
	Email              []string
	PreferredUsername  []string
	PhoneNumber        []string
	OAuthSubjectID     []string
	StandardAttributes map[string]interface{}

	Groups         []*Group
	EffectiveRoles []*Role
}

type ElasticsearchUserSource struct {
	ID          string     `json:"id,omitempty"`
	AppID       string     `json:"app_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	IsDisabled  bool       `json:"is_disabled"`

	Email     []string `json:"email,omitempty"`
	EmailText []string `json:"email_text,omitempty"`

	EmailLocalPart     []string `json:"email_local_part,omitempty"`
	EmailLocalPartText []string `json:"email_local_part_text,omitempty"`

	EmailDomain     []string `json:"email_domain,omitempty"`
	EmailDomainText []string `json:"email_domain_text,omitempty"`

	PreferredUsername     []string `json:"preferred_username,omitempty"`
	PreferredUsernameText []string `json:"preferred_username_text,omitempty"`

	PhoneNumber     []string `json:"phone_number,omitempty"`
	PhoneNumberText []string `json:"phone_number_text,omitempty"`

	PhoneNumberCountryCode []string `json:"phone_number_country_code,omitempty"`

	PhoneNumberNationalNumber     []string `json:"phone_number_national_number,omitempty"`
	PhoneNumberNationalNumberText []string `json:"phone_number_national_number_text,omitempty"`

	OAuthSubjectID     []string `json:"oauth_subject_id,omitempty"`
	OAuthSubjectIDText []string `json:"oauth_subject_id_text,omitempty"`

	FamilyName    string `json:"family_name,omitempty"`
	GivenName     string `json:"given_name,omitempty"`
	MiddleName    string `json:"middle_name,omitempty"`
	Name          string `json:"name,omitempty"`
	Nickname      string `json:"nickname,omitempty"`
	Gender        string `json:"gender,omitempty"`
	Zoneinfo      string `json:"zoneinfo,omitempty"`
	Locale        string `json:"locale,omitempty"`
	Formatted     string `json:"formatted,omitempty"`
	StreetAddress string `json:"street_address,omitempty"`
	Locality      string `json:"locality,omitempty"`
	Region        string `json:"region,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`

	RoleKey   []string `json:"role_key,omitempty"`
	RoleName  []string `json:"role_name,omitempty"`
	GroupKey  []string `json:"group_key,omitempty"`
	GroupName []string `json:"group_name,omitempty"`
}
