package model

import (
	"time"
)

type Session struct {
	Meta

	ACR string   `json:"acr,omitempty"`
	AMR []string `json:"amr,omitempty"`

	LastAccessedAt   time.Time `json:"last_accessed_at"`
	CreatedByIP      string    `json:"created_by_ip"`
	LastAccessedByIP string    `json:"last_accessed_by_ip"`
	UserAgent        UserAgent `json:"user_agent"`
}
