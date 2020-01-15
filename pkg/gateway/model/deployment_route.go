package model

import (
	"time"
)

const (
	DeploymentRouteTypeHTTPService string = "http-service"
	DeploymentRouteTypeStatic      string = "static"
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

func (r RouteTypeConfig) AssetPathMapping() map[string]string {
	m := r["asset_path_mapping"].(map[string]interface{})
	mapping := map[string]string{}
	for k, v := range m {
		mapping[k] = v.(string)
	}
	return mapping
}

func (r RouteTypeConfig) AssetFallbackPath() string {
	if p, ok := r["asset_fallback_path"].(string); ok {
		return p
	}
	return ""
}
