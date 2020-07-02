package handler

import "github.com/authgear/authgear-server/pkg/core/authn"

type AnonymousInteractionFlow interface {
	Authenticate(requestJWT string, clientID string) (*authn.Attrs, error)
}
