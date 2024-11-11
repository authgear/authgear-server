package facade

import (
	"context"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type AuthenticatorFacade struct {
	Coordinator *Coordinator
}

func (a AuthenticatorFacade) Get(ctx context.Context, id string) (*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorGet(ctx, id)
}

func (a AuthenticatorFacade) List(ctx context.Context, userID string, filters ...authenticator.Filter) ([]*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorList(ctx, userID, filters...)
}

func (a AuthenticatorFacade) New(ctx context.Context, spec *authenticator.Spec) (*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorNew(ctx, spec)
}

func (a AuthenticatorFacade) NewWithAuthenticatorID(ctx context.Context, authenticatorID string, spec *authenticator.Spec) (*authenticator.Info, error) {
	return a.Coordinator.AuthenticatorNewWithAuthenticatorID(ctx, authenticatorID, spec)
}

func (a AuthenticatorFacade) Create(ctx context.Context, authenticatorInfo *authenticator.Info, markVerified bool) error {
	return a.Coordinator.AuthenticatorCreate(ctx, authenticatorInfo, markVerified)
}

func (a AuthenticatorFacade) Update(ctx context.Context, authenticatorInfo *authenticator.Info) error {
	return a.Coordinator.AuthenticatorUpdate(ctx, authenticatorInfo)
}

func (a AuthenticatorFacade) UpdatePassword(ctx context.Context, authenticatorInfo *authenticator.Info, options *service.UpdatePasswordOptions) (changed bool, info *authenticator.Info, err error) {
	return a.Coordinator.AuthenticatorUpdatePassword(ctx, authenticatorInfo, options)
}

func (a AuthenticatorFacade) Delete(ctx context.Context, authenticatorInfo *authenticator.Info) error {
	return a.Coordinator.AuthenticatorDelete(ctx, authenticatorInfo)
}

func (a AuthenticatorFacade) VerifyWithSpec(ctx context.Context, info *authenticator.Info, spec *authenticator.Spec, options *VerifyOptions) (verifyResult *service.VerifyResult, err error) {
	return a.Coordinator.AuthenticatorVerifyWithSpec(ctx, info, spec, options)
}

func (a AuthenticatorFacade) VerifyOneWithSpec(ctx context.Context, userID string, authenticatorType apimodel.AuthenticatorType, infos []*authenticator.Info, spec *authenticator.Spec, options *VerifyOptions) (info *authenticator.Info, verifyResult *service.VerifyResult, err error) {
	return a.Coordinator.AuthenticatorVerifyOneWithSpec(ctx, userID, authenticatorType, infos, spec, options)
}

func (a AuthenticatorFacade) ClearLockoutAttempts(ctx context.Context, userID string, usedMethods []config.AuthenticationLockoutMethod) error {
	return a.Coordinator.AuthenticatorClearLockoutAttempts(ctx, userID, usedMethods)
}

func (a AuthenticatorFacade) MarkOOBIdentityVerified(ctx context.Context, info *authenticator.Info) error {
	return a.Coordinator.MarkOOBIdentityVerified(ctx, info)
}

type VerifyOptions struct {
	OOBChannel            *apimodel.AuthenticatorOOBChannel
	UseSubmittedValue     bool
	AuthenticationDetails *AuthenticationDetails
	Form                  otp.Form
}

func (v *VerifyOptions) toServiceOptions() *service.VerifyOptions {
	if v == nil {
		return nil
	}
	return &service.VerifyOptions{
		OOBChannel:        v.OOBChannel,
		UseSubmittedValue: v.UseSubmittedValue,
		Form:              v.Form,
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
