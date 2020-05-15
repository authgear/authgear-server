package authn

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Info struct {
	IsValid      bool
	UserID       string
	UserVerified bool
	UserDisabled bool

	SessionIdentityType   IdentityType
	SessionIdentityClaims map[string]interface{}
	SessionACR            string
	SessionAMR            []string
}

var _ Session = &Info{}

func NewAuthnInfo(attrs *Attrs, user *UserInfo) *Info {
	return &Info{
		IsValid:               true,
		UserID:                user.ID,
		UserVerified:          user.IsVerified,
		UserDisabled:          user.IsDisabled,
		SessionIdentityType:   attrs.IdentityType,
		SessionIdentityClaims: attrs.IdentityClaims,
		SessionACR:            attrs.ACR,
		SessionAMR:            attrs.AMR,
	}
}

const (
	headerSessionValid          = "X-Skygear-Session-Valid"
	headerUserID                = "X-Skygear-User-Id"
	headerUserVerified          = "X-Skygear-User-Verified"
	headerUserDisabled          = "X-Skygear-User-Disabled"
	headerSessionIdentityType   = "X-Skygear-Session-Identity-Type"
	headerSessionIdentityClaims = "X-Skygear-Session-Identity-Claims"
	headerSessionAcr            = "X-Skygear-Session-Acr"
	headerSessionAmr            = "X-Skygear-Session-Amr"
)

func (i *Info) PopulateHeaders(rw http.ResponseWriter) {
	if i == nil {
		return
	}

	rw.Header().Set(headerSessionValid, strconv.FormatBool(i.IsValid))
	if !i.IsValid {
		return
	}

	rw.Header().Set(headerUserID, i.UserID)
	rw.Header().Set(headerUserVerified, strconv.FormatBool(i.UserVerified))
	rw.Header().Set(headerUserDisabled, strconv.FormatBool(i.UserDisabled))

	rw.Header().Set(headerSessionIdentityType, string(i.SessionIdentityType))

	claimsJSON, err := json.Marshal(i.SessionIdentityClaims)
	if err != nil {
		panic(err)
	}
	claims := base64.RawURLEncoding.EncodeToString(claimsJSON)
	rw.Header().Set(headerSessionIdentityClaims, claims)

	rw.Header().Set(headerSessionAcr, i.SessionACR)
	rw.Header().Set(headerSessionAmr, strings.Join(i.SessionAMR, " "))
}

// TODO(authn): add session ID
func (i *Info) SessionID() string        { return "" }
func (i *Info) SessionType() SessionType { return SessionTypeAuthnInfo }

func (i *Info) AuthnAttrs() *Attrs {
	return &Attrs{
		UserID:       i.UserID,
		IdentityType: i.SessionIdentityType,
		ACR:          i.SessionACR,
		AMR:          i.SessionAMR,
	}
}

func (i *Info) User() *UserInfo {
	return &UserInfo{
		ID:         i.UserID,
		IsDisabled: i.UserDisabled,
		IsVerified: i.UserVerified,
	}
}

func ParseHeaders(r *http.Request) (*Info, error) {
	valid, err := strconv.ParseBool(r.Header.Get(headerSessionValid))
	if err != nil {
		return nil, nil
	}

	i := &Info{IsValid: valid}
	if !valid {
		return i, nil
	}

	i.UserID = r.Header.Get(headerUserID)
	if i.UserVerified, err = strconv.ParseBool(r.Header.Get(headerUserVerified)); err != nil {
		return nil, err
	}
	if i.UserDisabled, err = strconv.ParseBool(r.Header.Get(headerUserDisabled)); err != nil {
		return nil, err
	}

	i.SessionIdentityType = IdentityType(r.Header.Get(headerSessionIdentityType))

	claimsJSON, err := base64.RawURLEncoding.DecodeString(r.Header.Get(headerSessionIdentityClaims))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(claimsJSON, &i.SessionIdentityClaims); err != nil {
		return nil, err
	}

	i.SessionACR = r.Header.Get(headerSessionAcr)
	i.SessionAMR = strings.Split(r.Header.Get(headerSessionAmr), " ")

	return i, nil
}
