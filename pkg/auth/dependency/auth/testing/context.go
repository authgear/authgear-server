package testing

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type Builder struct {
	session *session.IDPSession
	user    *authn.UserInfo
}

func WithAuthn() Builder {
	return Builder{
		session: &session.IDPSession{
			ID: "session-id",
			Attrs: authn.Attrs{
				UserID: "user-id",
			},
		},
		user: &authn.UserInfo{
			ID: "user-id",
		},
	}
}

func (b Builder) ToRequest(r *http.Request) *http.Request {
	return r.WithContext(b.ToContext(r.Context()))
}

func (b Builder) ToContext(ctx context.Context) context.Context {
	return authn.WithAuthn(ctx, b.session, b.user)
}

func (b Builder) ToSession() auth.AuthSession {
	return b.session
}

func (b Builder) UserID(id string) Builder {
	b.user.ID = id
	b.session.Attrs.UserID = id
	return b
}

func (b Builder) SessionID(id string) Builder {
	b.session.ID = id
	return b
}
