package interaction

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
)

// Intent represents the intention that triggered the interaction.
type Intent interface {
	Type() IntentType
}

type IntentType string

const (
	IntentTypeOAuth               IntentType = "oauth"
	IntentTypeSignup              IntentType = "signup"
	IntentTypeLogin               IntentType = "login"
	IntentTypeAddIdentity         IntentType = "add-identity"
	IntentTypeRemoveIdentity      IntentType = "remove-identity"
	IntentTypeUpdateIdentity      IntentType = "update-identity"
	IntentTypeAddAuthenticator    IntentType = "add-authenticator"
	IntentTypeRemoveAuthenticator IntentType = "remove-authenticator"
	IntentTypeUpdateAuthenticator IntentType = "update-authenticator"
)

func NewIntent(t IntentType) Intent {
	switch t {
	case IntentTypeOAuth:
		return &IntentOAuth{}
	case IntentTypeSignup:
		return &IntentSignup{}
	case IntentTypeLogin:
		return &IntentLogin{}
	case IntentTypeAddIdentity:
		return &IntentAddIdentity{}
	case IntentTypeRemoveIdentity:
		return &IntentRemoveIdentity{}
	case IntentTypeUpdateIdentity:
		return &IntentUpdateIdentity{}
	case IntentTypeAddAuthenticator:
		return &IntentAddAuthenticator{}
	case IntentTypeRemoveAuthenticator:
		return &IntentRemoveAuthenticator{}
	}
	panic("interaction: unknown intent type " + t)
}

type OAuthAction string

const (
	OAuthActionLogin   OAuthAction = "login"
	OAuthActionLink    OAuthAction = "link"
	OAuthActionPromote OAuthAction = "promote"
)

type IntentOAuth struct {
	Identity                 identity.Spec `json:"identity"`
	Action                   OAuthAction   `json:"action"`
	Nonce                    string        `json:"nonce"`
	ProviderAuthorizationURL string        `json:"provider_authorization_url"`
	UserID                   string        `json:"user_id,omitempty"`
}

func (*IntentOAuth) Type() IntentType { return IntentTypeOAuth }

type IntentSignup struct {
	UserMetadata map[string]interface{} `json:"user_metadata"`
	Identity     identity.Spec          `json:"identity"`
}

func (*IntentSignup) Type() IntentType { return IntentTypeSignup }

type IntentLogin struct {
	Identity           identity.Spec `json:"identity"`
	OriginalIntentType IntentType    `json:"original_intent_type,omitempty"`
}

func (*IntentLogin) Type() IntentType { return IntentTypeLogin }

type IntentAddIdentity struct {
	Identity identity.Spec `json:"identity"`
}

func (*IntentAddIdentity) Type() IntentType { return IntentTypeAddIdentity }

type IntentRemoveIdentity struct {
	Identity identity.Spec `json:"identity"`
}

func (*IntentRemoveIdentity) Type() IntentType { return IntentTypeRemoveIdentity }

type IntentUpdateIdentity struct {
	OldIdentity identity.Spec `json:"old_identity"`
	NewIdentity identity.Spec `json:"new_identity"`
}

func (*IntentUpdateIdentity) Type() IntentType { return IntentTypeUpdateIdentity }

type IntentAddAuthenticator struct {
	Authenticator authenticator.Spec `json:"authenticator"`
	Secret        string             `json:"secret"`
}

func (*IntentAddAuthenticator) Type() IntentType { return IntentTypeAddAuthenticator }

type IntentRemoveAuthenticator struct {
	Authenticator authenticator.Spec `json:"authenticator"`
}

func (*IntentRemoveAuthenticator) Type() IntentType { return IntentTypeRemoveAuthenticator }

type IntentUpdateAuthenticator struct {
	Authenticator    authenticator.Spec `json:"authenticator"`
	OldSecret        string             `json:"-"`
	SkipVerifySecret bool               `json:"-"`
}

func (*IntentUpdateAuthenticator) Type() IntentType { return IntentTypeUpdateAuthenticator }
