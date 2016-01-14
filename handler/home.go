package handler

import (
	"github.com/oursky/skygear/router"
)

type statusResponse struct {
	Status string `json:"status,omitempty"`
}

// HomeHandler temp landing. FIXME
type HomeHandler struct {
}

func (h *HomeHandler) Setup() {
	return
}

func (h *HomeHandler) GetPreprocessors() []router.Processor {
	return nil
}

func (h *HomeHandler) Handle(playload *router.Payload, response *router.Response) {
	var (
		rep statusResponse
	)
	rep.Status = "OK"
	response.Result = rep
	return
}
