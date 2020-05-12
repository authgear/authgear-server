package interaction

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

// Intent represents the intention that triggered the interaction.
type Intent interface {
	Type() IntentType
}

type IntentType string

const (
	IntentTypeSignup              IntentType = "signup"
	IntentTypeLogin               IntentType = "login"
	IntentTypeAddIdentity         IntentType = "add-identity"
	IntentTypeRemoveIdentity      IntentType = "remove-identity"
	IntentTypeUpdateIdentity      IntentType = "update-identity"
	IntentTypeAddAuthenticator    IntentType = "add-authenticator"
	IntentTypeRemoveAuthenticator IntentType = "remove-authenticator"
)

func NewIntent(t IntentType) Intent {
	switch t {
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

type IntentSignup struct {
	UserMetadata    map[string]interface{} `json:"user_metadata"`
	Identity        identity.Spec          `json:"identity"`
	OnUserDuplicate model.OnUserDuplicate  `json:"on_user_duplicate"`
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
