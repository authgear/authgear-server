package model

import (
	"time"
)

type SessionType string

const (
	SessionTypeIDP          SessionType = "idp"
	SessionTypeOfflineGrant SessionType = "offline_grant"
)

type Session struct {
	Meta

	Type SessionType `json:"type"`

	AMR []string `json:"amr,omitempty"`

	LastAccessedAt                     time.Time `json:"lastAccessedAt"`
	CreatedByIP                        string    `json:"createdByIP"`
	LastAccessedByIP                   string    `json:"lastAccessedByIP"`
	LastAccessedByIPCountryCode        string    `json:"lastAccessedByIPCountryCode"`
	LastAccessedByIPEnglishCountryName string    `json:"lastAccessedByIPEnglishCountryName"`

	DisplayName     string `json:"displayName"`
	ApplicationName string `json:"applicationName,omitempty"`
}
