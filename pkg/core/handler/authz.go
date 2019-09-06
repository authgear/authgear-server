package handler

import (
	"encoding/json"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type authzMiddleware struct {
	authContext    auth.ContextGetter
	policyProvider authz.PolicyProvider
}

func (m authzMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// TODO(authz): proper logging
		log := logging.CreateLoggerWithContext(r.Context(), "server")

		policy := m.policyProvider.ProvideAuthzPolicy()
		if err := policy.IsAllowed(r, m.authContext); err != nil {
			log.WithError(err).Info("authz not allowed")
			m.writeUnauthorized(rw, err)
			return
		}

		next.ServeHTTP(rw, r)
	})
}

func (m authzMiddleware) writeUnauthorized(rw http.ResponseWriter, err error) {
	skyErr := skyerr.MakeError(err)
	httpStatus := skyerr.ErrorDefaultStatusCode(skyErr)
	response := APIResponse{Err: skyErr}
	encoder := json.NewEncoder(rw)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(httpStatus)
	encoder.Encode(response)
}

func RequireAuthz(h http.Handler, authCtx auth.ContextGetter, p authz.PolicyProvider) http.Handler {
	m := authzMiddleware{authContext: authCtx, policyProvider: p}
	return m.Handle(h)
}
