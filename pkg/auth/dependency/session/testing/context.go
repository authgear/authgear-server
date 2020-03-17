package testing

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

func WithSession(req *http.Request, userID string, principalID string) *http.Request {
	return req.WithContext(session.WithSession(
		req.Context(),
		&session.Session{
			Attrs: session.Attrs{
				UserID:      userID,
				PrincipalID: principalID,
			},
		},
		&authinfo.AuthInfo{
			ID: userID,
		},
	))
}
