package middleware

import (
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type PanicWriteEmptyResponseMiddleware struct{}

func (m *PanicWriteEmptyResponseMiddleware) Handle(next http.Handler) http.Handler {
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
				w.WriteHeader(apiError.Code)
				// Rethrow
				panic(err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
