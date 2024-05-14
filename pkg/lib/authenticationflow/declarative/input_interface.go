package declarative

import (
	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type syntheticInputOAuth interface {
	GetIdentitySpec() *identity.Spec
}

type inputTakeIdentificationMethod interface {
	GetIdentificationMethod() config.AuthenticationFlowIdentification
}

type inputTakeAccountRecoveryIdentificationMethod interface {
	GetAccountRecoveryIdentificationMethod() config.AuthenticationFlowAccountRecoveryIdentification
}

type inputTakeAccountRecoveryDestinationOptionIndex interface {
	GetAccountRecoveryDestinationOptionIndex() int
}

type inputTakeAccountLinkingIdentification interface {
	GetAccountLinkingIdentificationIndex() int
	GetAccountLinkingOAuthRedirectURI() string
	GetAccountLinkingOAuthResponseMode() string
}

type inputTakeAuthenticationMethod interface {
	GetAuthenticationMethod() config.AuthenticationFlowAuthentication
}

type inputTakeLoginID interface {
	GetLoginID() string
}

type inputTakeIDToken interface {
	GetIDToken() string
}

type inputTakeOAuthAuthorizationRequest interface {
	GetOAuthAlias() string
	GetOAuthRedirectURI() string
	GetOAuthResponseMode() string
	// We used to accept `state`.
	// But it turns out to be confusing.
	// `state` is used to maintain state between the request and the callback.
	// So ideally, `state` should contain the `state_token` of an authflow.
	// Therefore, we DO NOT accept `state`.
	// Instead, when the caller receive a authflow response that contains `oauth_authorization_url`,
	// they MUST add `state` to the URL, and then redirect.
}

type inputTakeOAuthAuthorizationResponse interface {
	GetOAuthAuthorizationCode() string
	GetOAuthError() string
	GetOAuthErrorDescription() string
	GetOAuthErrorURI() string
}

type inputTakePasskeyAssertionResponse interface {
	GetAssertionResponse() *protocol.CredentialAssertionResponse
}

type inputTakeOOBOTPChannel interface {
	GetChannel() model.AuthenticatorOOBChannel
}

type inputTakeOOBOTPTarget interface {
	GetTarget() string
}

type inputTakeNewPassword interface {
	GetNewPassword() string
}

type inputNodeVerifyClaim interface {
	IsCode() bool
	IsResend() bool
	IsCheck() bool
	GetCode() string
}

type inputNodeAuthenticationOOB interface {
	IsCode() bool
	IsResend() bool
	IsCheck() bool
	GetCode() string
}

type inputStepAccountRecoveryVerifyCode interface {
	IsCode() bool
	IsResend() bool
	GetCode() string
}

type inputSetupTOTP interface {
	GetCode() string
}

type inputConfirmRecoveryCode interface {
	ConfirmRecoveryCode()
}

type inputFillInUserProfile interface {
	GetAttributes() []attrs.T
}

type inputConfirmTerminateOtherSessions interface {
	ConfirmTerminateOtherSessions()
}

type inputTakePassword interface {
	GetPassword() string
}

type inputTakeAuthenticationOptionIndex interface {
	GetIndex() int
}

type inputTakeTOTP interface {
	GetCode() string
}

type inputTakeRecoveryCode interface {
	GetRecoveryCode() string
}

type inputNodePromptCreatePasskey interface {
	IsSkip() bool
	IsCreationResponse() bool
	GetCreationResponse() *protocol.CredentialCreationResponse
}
