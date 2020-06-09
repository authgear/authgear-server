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

	SessionACR string
	SessionAMR []string
}

var _ Session = &Info{}

func NewAuthnInfo(attrs *Attrs, user *UserInfo, isAnonymous bool) *Info {
	return &Info{
		IsValid:       true,
		UserID:        user.ID,
		UserAnonymous: isAnonymous,
		SessionACR:    attrs.ACR,
		SessionAMR:    attrs.AMR,
	}
}

const (
	headerSessionValid  = "X-Skygear-Session-Valid"
	headerUserID        = "X-Skygear-User-Id"
	headerUserAnonymous = "X-Skygear-User-Anonymous"
	headerSessionAcr    = "X-Skygear-Session-Acr"
	headerSessionAmr    = "X-Skygear-Session-Amr"
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

	rw.Header().Set(headerSessionAcr, i.SessionACR)
	rw.Header().Set(headerSessionAmr, strings.Join(i.SessionAMR, " "))
}

// TODO(authn): add session ID
func (i *Info) SessionID() string        { return "" }
func (i *Info) SessionType() SessionType { return SessionTypeAuthnInfo }

func (i *Info) AuthnAttrs() *Attrs {
	return &Attrs{
		UserID: i.UserID,
		ACR:    i.SessionACR,
		AMR:    i.SessionAMR,
	}
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

	i.SessionACR = r.Header.Get(headerSessionAcr)
	i.SessionAMR = strings.Split(r.Header.Get(headerSessionAmr), " ")

	return i, nil
}
