package testing

import (
	"context"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/session"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type Builder struct {
	session *session.IDPSession
}

func WithAuthn() Builder {
	return Builder{
		session: &session.IDPSession{
			ID: "session-id",
			Attrs: authn.Attrs{
				UserID: "user-id",
			},
		},
	}
}

func (b Builder) ToRequest(r *http.Request) *http.Request {
	return r.WithContext(b.ToContext(r.Context()))
}

func (b Builder) ToContext(ctx context.Context) context.Context {
	return authn.WithAuthn(ctx, b.session)
}

func (b Builder) ToSession() auth.AuthSession {
	return b.session
}

func (b Builder) UserID(id string) Builder {
	b.session.Attrs.UserID = id
	return b
}

func (b Builder) SessionID(id string) Builder {
	b.session.ID = id
	return b
}
