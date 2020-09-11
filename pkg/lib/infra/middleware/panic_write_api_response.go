package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type PanicWriteAPIResponseMiddleware struct{}

func (m *PanicWriteAPIResponseMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				var e error
				if ee, isErr := err.(error); isErr {
					e = ee
				} else {
					e = fmt.Errorf("%+v", err)
				}

				apiError := apierrors.AsAPIError(e)
				resp := &api.Response{Error: e}

				httpStatus := apiError.Code

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(httpStatus)
				encoder := json.NewEncoder(w)
				_ = encoder.Encode(resp)

				// Rethrow
				panic(err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
