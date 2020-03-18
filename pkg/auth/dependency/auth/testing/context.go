package testing

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type Builder struct {
	session *session.Session
	user    *authinfo.AuthInfo
}

func WithAuthn() Builder {
	return Builder{
		session: &session.Session{
			ID: "session-id",
			Attrs: authn.Attrs{
				UserID:      "user-id",
				PrincipalID: "principal-id",
			},
		},
		user: &authinfo.AuthInfo{
			ID:         "user-id",
			VerifyInfo: map[string]bool{},
		},
	}
}

func (b Builder) ToRequest(r *http.Request) *http.Request {
	return r.WithContext(b.ToContext(r.Context()))
}

func (b Builder) ToContext(ctx context.Context) context.Context {
	return authn.WithAuthn(ctx, b.session, b.user)
}

func (b Builder) UserID(id string) Builder {
	b.user.ID = id
	b.session.Attrs.UserID = id
	return b
}

func (b Builder) PrincipalID(id string) Builder {
	b.session.Attrs.PrincipalID = id
	return b
}

func (b Builder) SessionID(id string) Builder {
	b.session.ID = id
	return b
}

func (b Builder) Disabled(disabled bool) Builder {
	b.user.Disabled = disabled
	return b
}

func (b Builder) Verified(verified bool) Builder {
	b.user.Verified = verified
	return b
}

func (b Builder) VerifyInfo(info map[string]bool) Builder {
	b.user.VerifyInfo = info
	return b
}
