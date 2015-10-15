package handler

import (
	"github.com/oursky/skygear/router"
)

type statusResponse struct {
	Status string `json:"status,omitempty"`
}

// HomeHandler temp landing. FIXME
func HomeHandler(playload *router.Payload, response *router.Response) {
	var (
		rep statusResponse
	)
	rep.Status = "OK"
	response.Result = rep
	return
}
