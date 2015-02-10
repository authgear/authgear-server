package handlers

// Payloader is interface for passing payload to the actual handler
type Payloader interface {
	RouteAction() string
}

// Responser is interface for handler to write response to router
type Responser interface {
	Write([]byte) (int, error)
}
