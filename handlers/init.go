package handlers

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

// Responser is interface for handler to write response to router
type Response struct {
	Meta   map[string]interface{} `json:"-"`
	Result interface{}            `json:"result"`
}
