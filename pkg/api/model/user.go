package model

import (
	"time"
)

type User struct {
	Meta
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	IsAnonymous       bool       `json:"is_anonymous"`
	IsVerified        bool       `json:"is_verified"`
	IsDisabled        bool       `json:"is_disabled"`
	CanReauthenticate bool       `json:"can_reauthenticate"`
	DisableReason     *string    `json:"disable_reason"`
}

type ElasticsearchUserRaw struct {
	ID                string
	AppID             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	LastLoginAt       *time.Time
	IsDisabled        bool
	Email             []string
	PreferredUsername []string
	PhoneNumber       []string
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
}
