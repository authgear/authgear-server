package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type APIHandler interface {
	DecodeRequest(request *http.Request) (RequestPayload, error)
	Handle(requestPayload interface{}) (interface{}, error)
}

type APIResponse struct {
	Result interface{} `json:"result,omitempty"`
	Err    error       `json:"error,omitempty"`
}

func APIHandlerToHandler(apiHandler APIHandler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		payload, err := apiHandler.DecodeRequest(r)
		if err != nil {
			// TODO:
			// handle error properly
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		if err := payload.Validate(); err != nil {
			// TODO:
			// handle error properly
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		responsePayload, err := apiHandler.Handle(payload)
		response := APIResponse{}
		encoder := json.NewEncoder(rw)
		if err == nil {
			response.Result = responsePayload
		} else {
			// TODO:
			// update error handling
			response.Err = skyerr.MakeError(err)
		}

		rw.Header().Set("Content-Type", "application/json")
		encoder.Encode(response)
	})
}
