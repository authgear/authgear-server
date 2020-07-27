package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/otp"
)

func cloneAuthenticator(info *authenticator.Info) *authenticator.Info {
	newInfo := *info
	newInfo.Props = map[string]interface{}{}
	for k, v := range info.Props {
		newInfo.Props[k] = v
	}
	return &newInfo
}

func getAuthenticators(
	ctx *newinteraction.Context,
	graph *newinteraction.Graph,
	stage newinteraction.AuthenticationStage,
	typ authn.AuthenticatorType,
) (*identity.Info, []*authenticator.Info, error) {
	var identityInfo *identity.Info
	var infos []*authenticator.Info
	var err error
	if stage == newinteraction.AuthenticationStagePrimary {
		identityInfo = graph.MustGetUserLastIdentity()
		infos, err = ctx.Authenticators.ListByIdentity(identityInfo.UserID, identityInfo)

		n := 0
		for _, info := range infos {
			if info.Type == typ {
				infos[n] = info
				n++
			}
		}
		infos = infos[:n]
	} else {
		userID := graph.MustGetUserID()
		infos, err = ctx.Authenticators.List(userID, typ)
	}
	if err != nil {
		return nil, nil, err
	}

	return identityInfo, infos, nil
}

func sendOOBCode(
	ctx *newinteraction.Context,
	stage newinteraction.AuthenticationStage,
	operation otp.OOBOperationType,
	identityInfo *identity.Info,
	authenticatorInfo *authenticator.Info,
	secret string,
) error {
	if authenticatorInfo == nil {
		return nil
	}

	// TODO(interaction): handle rate limits

	channel := authn.AuthenticatorOOBChannel(authenticatorInfo.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string))

	var oobStage otp.OOBAuthenticationStage
	var loginID *loginid.LoginID
	if stage == newinteraction.AuthenticationStagePrimary {
		// Primary OOB authenticators is bound to login ID identities:
		// Extract login ID from the bound identity.
		oobStage = otp.OOBAuthenticationStagePrimary
		if identityInfo != nil {
			loginID = &loginid.LoginID{
				Key:   identityInfo.Claims[identity.IdentityClaimLoginIDKey].(string),
				Value: identityInfo.Claims[identity.IdentityClaimLoginIDValue].(string),
			}
		}
	} else {
		// Secondary OOB authenticators is not bound to login ID identities.
		oobStage = otp.OOBAuthenticationStageSecondary
		loginID = nil
	}

	// Use a placeholder login ID if no bound login ID identity
	if loginID == nil {
		loginID = &loginid.LoginID{}
		switch channel {
		case authn.AuthenticatorOOBChannelSMS:
			loginID.Value = authenticatorInfo.Props[authenticator.AuthenticatorPropOOBOTPPhone].(string)
		case authn.AuthenticatorOOBChannelEmail:
			loginID.Value = authenticatorInfo.Props[authenticator.AuthenticatorPropOOBOTPEmail].(string)
		}
	}

	code := ctx.OOBAuthenticators.GenerateCode(secret, channel)
	return ctx.OOBAuthenticators.SendCode(channel, loginID, code, operation, oobStage)
}
