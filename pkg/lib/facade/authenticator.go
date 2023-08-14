package facade

import (
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AuthenticatorFacade struct {
	Coordinator *Coordinator
}

func (a AuthenticatorFacade) Get(id string) (*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorGet(id)
}

func (a AuthenticatorFacade) List(userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorList(userID, filters...)
}

func (a AuthenticatorFacade) New(spec *authenticator.Spec) (*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorNew(spec)
}

func (a AuthenticatorFacade) NewWithAuthenticatorID(authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorNewWithAuthenticatorID(authenticatorID, spec)
}

func (a AuthenticatorFacade) WithSpec(authenticatorInfo *authenticator.Info, spec *authenticator.Spec) (changed bool, info *authenticator.Info, err error) {
	return a.Coordinator.AuthenticatorWithSpec(authenticatorInfo, spec)
}

func (a AuthenticatorFacade) Create(authenticatorInfo *authenticator.Info, markVerified bool) error {
	return a.Coordinator.AuthenticatorCreate(authenticatorInfo, markVerified)
}

func (a AuthenticatorFacade) Update(authenticatorInfo *authenticator.Info) error {
	return a.Coordinator.AuthenticatorUpdate(authenticatorInfo)
}

func (a AuthenticatorFacade) Delete(authenticatorInfo *authenticator.Info) error {
	return a.Coordinator.AuthenticatorDelete(authenticatorInfo)
}

func (a AuthenticatorFacade) VerifyWithSpec(info *authenticator.Info, spec *authenticator.Spec, options *VerifyOptions) (requireUpdate bool, err error) {
	return a.Coordinator.AuthenticatorVerifyWithSpec(info, spec, options)
}

func (a AuthenticatorFacade) VerifyOneWithSpec(infos []*authenticator.Info, spec *authenticator.Spec, options *VerifyOptions) (info *authenticator.Info, requireUpdate bool, err error) {
	return a.Coordinator.AuthenticatorVerifyOneWithSpec(infos, spec, options)
}

func (a AuthenticatorFacade) ClearLockoutAttempts(userID string, usedMethods []config.AuthenticationLockoutMethod) error {
	return a.Coordinator.AuthenticatorClearLockoutAttempts(userID, usedMethods)
}

type VerifyOptions struct {
	OOBChannel            *apimodel.AuthenticatorOOBChannel
	AuthenticationDetails *AuthenticationDetails
}

func (v *VerifyOptions) toServiceOptions() *service.VerifyOptions {
	if v == nil {
		return nil
	}
	return &service.VerifyOptions{
		OOBChannel: v.OOBChannel,
	}
}

type AuthenticationDetails struct {
	UserID             string
	Stage              authn.AuthenticationStage
	AuthenticationType authn.AuthenticationType
}

func NewAuthenticationDetails(
	userID string,
	stage authn.AuthenticationStage,
	authenticationType authn.AuthenticationType,
) *AuthenticationDetails {
	return &AuthenticationDetails{
		UserID:             userID,
		Stage:              stage,
		AuthenticationType: authenticationType,
	}
}
