package authn

type ClaimName string

// ref: https://www.iana.org/assignments/jwt/jwt.xhtml
const (
	ClaimACR               ClaimName = "acr"
	ClaimAMR               ClaimName = "amr"
	ClaimEmail             ClaimName = "email"
	ClaimPhoneNumber       ClaimName = "phone_number"
	ClaimPreferredUsername ClaimName = "preferred_username"
	ClaimKeyID             ClaimName = "https://authgear.com/user/key_id"
	ClaimUserIsAnonymous   ClaimName = "https://authgear.com/user/is_anonymous"
	ClaimUserIsVerified    ClaimName = "https://authgear.com/user/is_verified"
	ClaimUserMetadata      ClaimName = "https://authgear.com/user/metadata"
)

type Attrs struct {
	UserID string                    `json:"user_id"`
	Claims map[ClaimName]interface{} `json:"claims"`
}

func (a *Attrs) GetACR() (string, bool) {
	acr, ok := a.Claims[ClaimACR].(string)
	return acr, ok
}

func (a *Attrs) SetACR(value string) {
	if len(value) > 0 {
		a.Claims[ClaimACR] = value
	} else {
		delete(a.Claims, ClaimACR)
	}
}

func (a *Attrs) GetAMR() ([]string, bool) {
	amr, ok := a.Claims[ClaimAMR].([]string)
	return amr, ok
}

func (a *Attrs) SetAMR(value []string) {
	if len(value) > 0 {
		a.Claims[ClaimAMR] = value
	} else {
		delete(a.Claims, ClaimAMR)
	}
}

func (a *Attrs) AuthnAttrs() *Attrs {
	return a
}

type Attributer interface {
	AuthnAttrs() *Attrs
}
