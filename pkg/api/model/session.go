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

	ACR string   `json:"acr,omitempty"`
	AMR []string `json:"amr,omitempty"`

	LastAccessedAt   time.Time `json:"lastAccessedAt"`
	CreatedByIP      string    `json:"createdByIP"`
	LastAccessedByIP string    `json:"lastAccessedByIP"`
}
