package model

import (
	"time"
)

const (
	DeploymentRouteTypeHTTPService string = "http-service"
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

func (r RouteTypeConfig) BackendURL() string {
	if str, ok := r["backend_url"].(string); ok {
		return str
	}
	return ""
}

func (r RouteTypeConfig) TargetPath() string {
	if str, ok := r["target_path"].(string); ok {
		return str
	}
	return ""
}
