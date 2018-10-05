package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/handler/context"
	nextSkyerr "github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type APIHandler interface {
	DecodeRequest(request *http.Request) (RequestPayload, error)
	Handle(requestPayload interface{}, ctx context.AuthContext) (interface{}, error)
}

type APIResponse struct {
	Result interface{}  `json:"result,omitempty"`
	Err    skyerr.Error `json:"error,omitempty"`
}

func APIHandlerToHandler(apiHandler APIHandler) Handler {
	return HandlerFunc(func(rw http.ResponseWriter, r *http.Request, ctx context.AuthContext) {
		httpStatus := http.StatusOK
		response := APIResponse{}
		encoder := json.NewEncoder(rw)

		defer func() {
			if response.Err != nil {
				httpStatus = nextSkyerr.ErrorDefaultStatusCode(response.Err)
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(httpStatus)
			encoder.Encode(response)
		}()

		payload, err := apiHandler.DecodeRequest(r)
		if err != nil {
			response.Err = skyerr.MakeError(err)
			return
		}

		if err := payload.Validate(); err != nil {
			response.Err = skyerr.MakeError(err)
			return
		}

		responsePayload, err := apiHandler.Handle(payload, ctx)
		if err == nil {
			response.Result = responsePayload
		} else {
			response.Err = skyerr.MakeError(err)
		}
	})
}
