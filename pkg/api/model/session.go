package model

import (
	"time"
)

type Session struct {
	Meta

	ACR string   `json:"acr,omitempty"`
	AMR []string `json:"amr,omitempty"`

	LastAccessedAt   time.Time `json:"lastAccessedAt"`
	CreatedByIP      string    `json:"createdByIP"`
	LastAccessedByIP string    `json:"lastAccessedByIP"`
	UserAgent        UserAgent `json:"userAgent"`
}
