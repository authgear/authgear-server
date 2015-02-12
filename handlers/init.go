package handlers

// Payload is for passing payload to the actual handler
type Payload struct {
	Data map[string]interface{}
	Raw []byte
}

// RouteAction must exist for every request
func (p *Payload) RouteAction() string {
	return p.Data["action"].(string)
}

// Responser is interface for handler to write response to router
type Responser interface {
	Write([]byte) (int, error)
}
