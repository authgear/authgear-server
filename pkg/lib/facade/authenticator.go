package facade

import (
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

type AuthenticatorFacade struct {
	Coordinator *Coordinator
}

func (a AuthenticatorFacade) Get(userID string, typ authn.AuthenticatorType, id string) (*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorGet(userID, typ, id)
}

func (a AuthenticatorFacade) List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorList(userID, filters...)
}

func (a AuthenticatorFacade) New(spec *authenticator.Spec, secret string) (*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorNew(spec, secret)
}

func (a AuthenticatorFacade) WithSecret(authenticatorInfo *authenticator.Info, secret string) (changed bool, info *authenticator.Info, err error) {
	return a.Coordinator.AuthenticatorWithSecret(authenticatorInfo, secret)
}

func (a AuthenticatorFacade) Create(authenticatorInfo *authenticator.Info) error {
	return a.Coordinator.AuthenticatorCreate(authenticatorInfo)
}

func (a AuthenticatorFacade) Update(authenticatorInfo *authenticator.Info) error {
	return a.Coordinator.AuthenticatorUpdate(authenticatorInfo)
}

func (a AuthenticatorFacade) Delete(authenticatorInfo *authenticator.Info) error {
	return a.Coordinator.AuthenticatorDelete(authenticatorInfo)
}

func (a AuthenticatorFacade) VerifySecret(info *authenticator.Info, secret string) (requireUpdate bool, err error) {
	return a.Coordinator.AuthenticatorVerifySecret(info, secret)
}
