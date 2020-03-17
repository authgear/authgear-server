package testing

import (
	"context"
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type session struct {
	ID string
	authn.Attrs
}

func (s *session) SessionID() string              { return s.ID }
func (s *session) SessionType() authn.SessionType { return authn.SessionTypeIdentityProvider }

func (s *session) AuthnAttrs() *authn.Attrs {
	return &s.Attrs
}

type Builder struct {
	sessionID string
	attrs     authn.Attrs
	user      *authinfo.AuthInfo
}

func WithAuthn() Builder {
	return Builder{
		sessionID: "session-id",
		attrs: authn.Attrs{
			UserID:      "user-id",
			PrincipalID: "principal-id",
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
	return authn.WithAuthn(ctx,
		&session{
			ID:    b.sessionID,
			Attrs: b.attrs,
		},
		b.user,
	)
}

func (b Builder) UserID(id string) Builder {
	b.user.ID = id
	b.attrs.UserID = id
	return b
}

func (b Builder) PrincipalID(id string) Builder {
	b.attrs.PrincipalID = id
	return b
}

func (b Builder) SessionID(id string) Builder {
	b.sessionID = id
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
