package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	nextSkyerr "github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type APIHandler interface {
	DecodeRequest(request *http.Request) (RequestPayload, error)
	WithTx() bool
	Handle(requestPayload interface{}) (interface{}, error)
}

type APIResponse struct {
	Result interface{}  `json:"result,omitempty"`
	Err    skyerr.Error `json:"error,omitempty"`
}

func APIHandlerToHandler(apiHandler APIHandler, txContext db.TxContext) http.Handler {
	handleAPICall := func(r *http.Request) (response APIResponse) {
		payload, err := apiHandler.DecodeRequest(r)
		if err != nil {
			response.Err = skyerr.MakeError(err)
			return
		}

		if err := payload.Validate(); err != nil {
			response.Err = skyerr.MakeError(err)
			return
		}

		if apiHandler.WithTx() {
			// assume txContext != nil if apiHandler.WithTx() is true
			if err := txContext.BeginTx(); err != nil {
				panic(err)
			}

			defer func() {
				if txContext.HasTx() {
					txContext.RollbackTx()
				}
			}()
		}

		responsePayload, err := apiHandler.Handle(payload)

		if err == nil {
			response.Result = responsePayload

			if txContext != nil {
				txContext.CommitTx()
			}
		} else {
			response.Err = skyerr.MakeError(err)
		}

		return
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		response := handleAPICall(r)
		WriteResponse(rw, response)
	})
}

func WriteResponse(rw http.ResponseWriter, response APIResponse) {
	httpStatus := http.StatusOK
	encoder := json.NewEncoder(rw)

	if response.Err != nil {
		httpStatus = nextSkyerr.ErrorDefaultStatusCode(response.Err)
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(httpStatus)
	encoder.Encode(response)
}

// Transactional runs f within a transaction.
// If err is non-nil, the transaction is rolled back.
// Otherwise the transaction is committed.
// It is a lightweight and flexible alternative to APIHandler
// because it is not coupled with http.
func Transactional(txContext db.TxContext, f func() (interface{}, error)) (interface{}, error) {
	if e := txContext.BeginTx(); e != nil {
		panic(e)
	}
	defer func() {
		if txContext.HasTx() {
			txContext.RollbackTx()
		}
	}()
	result, err := f()
	if err != nil {
		if e := txContext.RollbackTx(); e != nil {
			panic(e)
		}
	} else {
		if e := txContext.CommitTx(); e != nil {
			panic(e)
		}
	}
	return result, err
}
