package model

import (
	"time"
)

type DeploymentRoute struct {
	ID         string
	CreatedAt  *time.Time
	Version    string
	Path       string
	Type       string
	TypeConfig RouteTypeConfig
}

type RouteTypeConfig map[string]interface{}
