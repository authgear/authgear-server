package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
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
		infos, err = ctx.Authenticators.ListAll(identityInfo.UserID)
		if err != nil {
			return nil, nil, err
		}
		infos = ctx.Authenticators.FilterPrimaryAuthenticators(identityInfo, infos)

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
	isAuthenticating bool,
	identityInfo *identity.Info,
	authenticatorInfo *authenticator.Info,
	secret string,
) (*otp.OOBSendResult, error) {
	// TODO(interaction): handle rate limits

	channel := authn.AuthenticatorOOBChannel(authenticatorInfo.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string))

	var target string
	var messageType otp.MessageType
	if stage == newinteraction.AuthenticationStagePrimary {
		// Primary OOB authenticators should match login ID identities:
		// Extract login ID from the identity.
		if identityInfo != nil {
			target = identityInfo.Claims[identity.IdentityClaimLoginIDValue].(string)
		}

		if isAuthenticating {
			messageType = otp.MessageTypeAuthenticatePrimaryOOB
		} else {
			messageType = otp.MessageTypeSetupPrimaryOOB
		}
	} else {
		// Secondary OOB authenticators is not related to login ID identities.
		if isAuthenticating {
			messageType = otp.MessageTypeAuthenticateSecondaryOOB
		} else {
			messageType = otp.MessageTypeSetupSecondaryOOB
		}
	}

	// Use a placeholder login ID if no matching login ID identity
	if target == "" {
		switch channel {
		case authn.AuthenticatorOOBChannelSMS:
			target = authenticatorInfo.Props[authenticator.AuthenticatorPropOOBOTPPhone].(string)
		case authn.AuthenticatorOOBChannelEmail:
			target = authenticatorInfo.Props[authenticator.AuthenticatorPropOOBOTPEmail].(string)
		}
	}

	code := ctx.OOBAuthenticators.GenerateCode(secret, channel)
	return ctx.OOBAuthenticators.SendCode(channel, target, code, messageType)
}

func stageToAuthenticatorTag(stage newinteraction.AuthenticationStage) []string {
	switch stage {
	case newinteraction.AuthenticationStagePrimary:
		return []string{authenticator.TagPrimaryAuthenticator}
	case newinteraction.AuthenticationStageSecondary:
		return []string{authenticator.TagSecondaryAuthenticator}
	default:
		panic("interaction: unknown stage: " + stage)
	}
}
