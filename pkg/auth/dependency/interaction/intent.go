package interaction

// Intent represents the intention that triggered the interaction.
type Intent interface {
	intentType() string
}

type IntentSignup struct {
	UserMetadata map[string]interface{} `json:"user_metadata"`
	Identity     IdentitySpec           `json:"identity"`
}

func (IntentSignup) intentType() string { return "signup" }

type IntentLogin struct {
	Identity IdentitySpec `json:"identity"`
}

func (IntentLogin) intentType() string { return "login" }

type IntentAddIdentity struct {
	Identity IdentitySpec `json:"identity"`
}

func (IntentAddIdentity) intentType() string { return "add-identity" }

type IntentRemoveIdentity struct {
	Identity IdentitySpec `json:"identity"`
}

func (IntentRemoveIdentity) intentType() string { return "remove-identity" }

type IntentUpdateIdentity struct {
	OldIdentity IdentitySpec `json:"old_identity"`
	NewIdentity IdentitySpec `json:"new_identity"`
}

func (IntentUpdateIdentity) intentType() string { return "update-identity" }

type IntentAddAuthenticator struct {
	Authenticator AuthenticatorSpec `json:"authenticator"`
	Secret        string            `json:"secret"`
}

func (IntentAddAuthenticator) intentType() string { return "add-authenticator" }

type IntentRemoveAuthenticator struct {
	Authenticator AuthenticatorSpec `json:"authenticator"`
}

func (IntentRemoveAuthenticator) intentType() string { return "remove-authenticator" }
