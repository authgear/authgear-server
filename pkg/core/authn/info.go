package authn

import (
	"net/http"
	"strconv"
	"strings"
)

type Info struct {
	IsValid       bool
	UserID        string
	UserAnonymous bool
	UserVerified  bool

	SessionACR string
	SessionAMR []string
}

var _ Session = &Info{}

func NewAuthnInfo(attrs *Attrs, isAnonymous bool, isVerified bool) *Info {
	acr, _ := attrs.GetACR()
	amr, _ := attrs.GetAMR()
	return &Info{
		IsValid:       true,
		UserID:        attrs.UserID,
		UserAnonymous: isAnonymous,
		UserVerified:  isVerified,
		SessionACR:    acr,
		SessionAMR:    amr,
	}
}

const (
	headerSessionValid  = "X-Authgear-Session-Valid"
	headerUserID        = "X-Authgear-User-Id"
	headerUserVerified  = "X-Authgear-User-Verified"
	headerUserAnonymous = "X-Authgear-User-Anonymous"
	headerSessionAcr    = "X-Authgear-Session-Acr"
	headerSessionAmr    = "X-Authgear-Session-Amr"
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
	rw.Header().Set(headerUserAnonymous, strconv.FormatBool(i.UserAnonymous))
	rw.Header().Set(headerUserVerified, strconv.FormatBool(i.UserVerified))

	rw.Header().Set(headerSessionAcr, i.SessionACR)
	rw.Header().Set(headerSessionAmr, strings.Join(i.SessionAMR, " "))
}

// TODO(authn): add session ID
func (i *Info) SessionID() string        { return "" }
func (i *Info) SessionType() SessionType { return SessionTypeAuthnInfo }

func (i *Info) AuthnAttrs() *Attrs {
	attrs := &Attrs{
		UserID: i.UserID,
		Claims: map[ClaimName]interface{}{},
	}
	attrs.SetACR(i.SessionACR)
	attrs.SetAMR(i.SessionAMR)
	return attrs
}

func (i *Info) User() *UserInfo {
	return &UserInfo{
		ID: i.UserID,
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
	if i.UserAnonymous, err = strconv.ParseBool(r.Header.Get(headerUserAnonymous)); err != nil {
		return nil, err
	}
	if i.UserVerified, err = strconv.ParseBool(r.Header.Get(headerUserVerified)); err != nil {
		return nil, err
	}

	i.SessionACR = r.Header.Get(headerSessionAcr)
	i.SessionAMR = strings.Split(r.Header.Get(headerSessionAmr), " ")

	return i, nil
}
