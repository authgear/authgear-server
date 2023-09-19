package declarative

import (
	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type inputTakeIdentificationMethod interface {
	GetIdentificationMethod() config.AuthenticationFlowIdentification
}

type inputTakeAuthenticationMethod interface {
	GetAuthenticationMethod() config.AuthenticationFlowAuthentication
}

type inputTakeLoginID interface {
	GetLoginID() string
}

type inputTakeOAuthAuthorizationRequest interface {
	GetOAuthAlias() string
	GetOAuthState() string
	GetOAuthRedirectURI() string
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

type inputSetupTOTP interface {
	GetCode() string
	GetDisplayName() string
}

type inputConfirmRecoveryCode interface {
	ConfirmRecoveryCode()
}

type inputFillUserProfile interface {
	GetAttributes() []attrs.T
}

type inputConfirmTerminateOtherSessions interface {
	ConfirmTerminateOtherSessions()
}

type inputTakePassword interface {
	GetPassword() string
}

type inputTakeAuthenticationCandidateIndex interface {
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
