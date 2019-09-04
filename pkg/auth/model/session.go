package model

import (
	"time"
)

// Session is the API model of user sessions
type Session struct {
	ID               string                 `json:"id"`
	IdentityID       string                 `json:"identity_id"`
	CreatedAt        time.Time              `json:"created_at"`
	LastAccessedAt   time.Time              `json:"last_accessed_at"`
	CreatedByIP      string                 `json:"created_by_ip"`
	LastAccessedByIP string                 `json:"last_accessed_by_ip"`
	UserAgent        SessionUserAgent       `json:"user_agent"`
	Name             string                 `json:"name"`
	Data             map[string]interface{} `json:"data"`
}

// Session is the API model of user agent of session
type SessionUserAgent struct {
	Raw         string `json:"raw"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	OS          string `json:"os"`
	OSVersion   string `json:"os_version"`
	DeviceName  string `json:"device_name"`
	DeviceModel string `json:"device_model"`
}

// @JSONSchema
const SessionSchema = `
{
	"$id": "#Session",
	"type": "object",
	"properties": {
		"id": { "type": "string" },
		"identity_id": { "type": "string" },
		"created_at": { "type": "string" },
		"last_accessed_at": { "type": "string" },
		"created_by_ip": { "type": "string" },
		"last_accessed_by_ip": { "type": "string" },
		"user_agent": { "$ref": "#SessionUserAgent" },
		"name": { "type": "string" },
		"data": { "type": "object" }
	}
}
`

// @JSONSchema
const SessionUserAgentSchema = `
{
	"$id": "#SessionUserAgent",
	"type": "object",
	"properties": {
		"raw": { "type": "string" },
		"name": { "type": "string" },
		"version": { "type": "string" },
		"os": { "type": "string" },
		"os_version": { "type": "string" },
		"device_name": { "type": "string" },
		"device_model": { "type": "string" }
	}
}
`

// @JSONSchema
const SessionResponseSchema = `
{
	"$id": "#SessionResponse",
	"type": "object",
	"properties": {
		"result": { "$ref": "#Session" }
	}
}
`
