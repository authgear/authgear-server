package server

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	nextSkyerr "github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type Option struct {
	RecoverPanic        bool
	RecoverPanicHandler middleware.RecoverHandler
	GearPathPrefix      string
}

// RecoveredResponse is interface for the default RecoverPanicHandler to write response
type RecoveredResponse struct {
	Err skyerr.Error `json:"error,omitempty"`
}

func DefaultRecoverPanicHandler(w http.ResponseWriter, r *http.Request, err skyerr.Error) {
	httpStatus := nextSkyerr.ErrorDefaultStatusCode(err)

	// TODO: log

	response := RecoveredResponse{Err: err}
	encoder := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	encoder.Encode(response)
}

func DefaultOption() Option {
	return Option{
		RecoverPanic:        true,
		RecoverPanicHandler: DefaultRecoverPanicHandler,
		GearPathPrefix:      "",
	}
}
