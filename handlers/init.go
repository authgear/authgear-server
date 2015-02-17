package handlers

import (
	"log"
)

// Payload is for passing payload to the actual handler
type Payload struct {
	// Map of params such as Auth, TimeSteam, version
	Meta map[string]interface{}
	// Map of action payload
	Data map[string]interface{}
}

// RouteAction must exist for every request
func (p *Payload) RouteAction() string {
	return p.Data["action"].(string)
}

// IsAuth tell the middleware is this payload is an auth request
func (p *Payload) IsAuth() bool {
	defer func() {
		if r := recover(); r != nil {
			log.Println("IsAuth recover")
		}
		return
	}()
	return p.Data["action"].(string)[0:5] == "auth:"
}

// AccessToken return the user input string
// TODO: accept all header, json payload, query string(in order)
func (p *Payload) AccessToken() string {
	var token interface{}
	token = p.Data["access_token"]
	switch token := token.(type) {
	default:
		return ""
	case string:
		return token
	}
}

// Response is interface for handler to write response to router
type Response struct {
	Meta       map[string]interface{} `json:"-"`
	Result     interface{}            `json:"result"`
	RequestID  string                 `json:"request_id,omitempty"`
	DatabaseID string                 `json:"database_id,omitempty"`
}
