package session

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

func NewInfo(attrs *Attrs, isAnonymous bool, isVerified bool) *Info {
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
