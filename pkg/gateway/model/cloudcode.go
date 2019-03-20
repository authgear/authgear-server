package model

import (
	"time"
)

type CloudCode struct {
	ID         string
	CreatedAt  *time.Time
	Version    string
	Path       string
	TargetPath string
	BackendURL string
}
