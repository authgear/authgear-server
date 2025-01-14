package oauth

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/session"
)

type GrantSessionKind string

const (
	GrantSessionKindOffline GrantSessionKind = "offline_grant"
	GrantSessionKindSession GrantSessionKind = "idp_session"
)

func (k GrantSessionKind) SessionType() session.Type {
	switch k {
	case GrantSessionKindSession:
		return session.TypeIdentityProvider
	case GrantSessionKindOffline:
		return session.TypeOfflineGrant
	default:
		panic(fmt.Errorf("unknown session kind: %v\n", k))
	}
}

func GrantSessionKindFromSessionType(typ session.Type) GrantSessionKind {
	switch typ {
	case session.TypeIdentityProvider:
		return GrantSessionKindSession
	case session.TypeOfflineGrant:
		return GrantSessionKindOffline
	default:
		panic(fmt.Errorf("unknown session type: %v", typ))
	}
}
