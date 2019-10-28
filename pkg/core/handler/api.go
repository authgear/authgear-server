package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type APIHandler interface {
	DecodeRequest(request *http.Request, resp http.ResponseWriter) (RequestPayload, error)
	WithTx() bool
	Handle(requestPayload interface{}) (interface{}, error)
}

type APITxHandler interface {
	WillCommitTx() error
	DidCommitTx()
}

type APIResponse struct {
	Result interface{}
	Error  error
}

func (r APIResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Result interface{}      `json:"result,omitempty"`
		Error  *skyerr.APIError `json:"error,omitempty"`
	}{r.Result, skyerr.AsAPIError(r.Error)})
}

func APIHandlerToHandler(apiHandler APIHandler, txContext db.TxContext) http.Handler {
	txHandler, _ := apiHandler.(APITxHandler)

	handleAPICall := func(r *http.Request, resp http.ResponseWriter) (response APIResponse) {
		payload, err := apiHandler.DecodeRequest(r, resp)
		if err != nil {
			response.Error = err
			return
		}

		if err := payload.Validate(); err != nil {
			response.Error = err
			return
		}

		defer func() {
			if err != nil {
				response.Error = err
			}
		}()

		if apiHandler.WithTx() {
			// assume txContext != nil if apiHandler.WithTx() is true
			if err := txContext.BeginTx(); err != nil {
				response.Error = err
				return
			}

			defer func() {
				err = db.EndTx(txContext, err)
				if err == nil && txHandler != nil {
					txHandler.DidCommitTx()
				}
			}()
		}

		responsePayload, err := apiHandler.Handle(payload)

		if err == nil && txHandler != nil {
			err = txHandler.WillCommitTx()
		}

		if err == nil {
			response.Result = responsePayload
		}

		return
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		response := handleAPICall(r, rw)
		WriteResponse(rw, response)
	})
}

// HandledError represents a handled (i.e. API responded with error) unexpected
// error. When encountered this error, panic recovery middleware should log
// the error, without changing the response.
type HandledError struct {
	Error error
}

func WriteResponse(rw http.ResponseWriter, response APIResponse) {
	httpStatus := http.StatusOK
	encoder := json.NewEncoder(rw)
	err := skyerr.AsAPIError(response.Error)

	if err != nil {
		httpStatus = err.Code
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(httpStatus)
	encoder.Encode(response)

	if err != nil && err.Code >= 500 && err.Code < 600 {
		// delegate logging to panic recovery
		panic(HandledError{response.Error})
	}
}

// Transactional runs f within a transaction.
// If err is non-nil, the transaction is rolled back.
// Otherwise the transaction is committed.
// It is a lightweight and flexible alternative to APIHandler
// because it is not coupled with http.
func Transactional(txContext db.TxContext, f func() (interface{}, error)) (result interface{}, err error) {
	err = db.WithTx(txContext, func() error {
		result, err = f()
		return err
	})
	return
}
