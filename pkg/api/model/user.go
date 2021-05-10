package model

import (
	"time"
)

type User struct {
	Meta
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	IsAnonymous   bool       `json:"is_anonymous"`
	IsVerified    bool       `json:"is_verified"`
	IsDisabled    bool       `json:"is_disabled"`
	DisableReason *string    `json:"disable_reason"`
}

type ElasticsearchUser struct {
	ID          string     `json:"id,omitempty"`
	AppID       string     `json:"app_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	IsDisabled  bool       `json:"is_disabled"`

	Email          []string `json:"email,omitempty"`
	EmailLocalPart []string `json:"email_local_part,omitempty"`
	EmailDomain    []string `json:"email_domain,omitempty"`

	PreferredUsername []string `json:"preferred_username,omitempty"`

	PhoneNumber               []string `json:"phone_number,omitempty"`
	PhoneNumberCountryCode    []string `json:"phone_number_country_code,omitempty"`
	PhoneNumberNationalNumber []string `json:"phone_number_national_number,omitempty"`
}
