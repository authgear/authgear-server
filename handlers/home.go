package handlers

import ()

type statusResponse struct {
	Status string `json:"status,omitempty"`
}

// HomeHandler temp landing. FIXME
func HomeHandler(playload Payload) Response {
	var (
		response Response
		rep      statusResponse
	)
	rep.Status = "OK"
	response.Result = rep
	return response
}
