package handlers

import ()

type statusResponse struct {
	Status string `json:"status,omitempty"`
}

// HomeHandler temp landing. FIXME
func HomeHandler(playload *Payload, response *Response) {
	var (
		rep statusResponse
	)
	rep.Status = "OK"
	response.Result = rep
	return
}
