package handler

import "github.com/skygeario/skygear-server/pkg/core/authn"

type AnonymousInteractionFlow interface {
	Authenticate(requestJWT string, clientID string) (*authn.Attrs, error)
}
