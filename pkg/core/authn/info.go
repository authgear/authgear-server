package authn

import (
	"net/http"
	"strconv"
	"time"
)

type Info struct {
	IsValid      bool
	UserID       string
	UserVerified bool
	UserDisabled bool

	SessionIdentityID        string
	SessionIdentityType      PrincipalType
	SessionIdentityUpdatedAt time.Time

	SessionAuthenticatorID         string
	SessionAuthenticatorType       AuthenticatorType
	SessionAuthenticatorOOBChannel AuthenticatorOOBChannel
	SessionAuthenticatorUpdatedAt  *time.Time
}

const (
	headerSessionValid                   = "X-Skygear-Session-Valid"
	headerUserID                         = "X-Skygear-User-Id"
	headerUserVerified                   = "X-Skygear-User-Verified"
	headerUserDisabled                   = "X-Skygear-User-Disabled"
	headerSessionIdentityID              = "X-Skygear-Session-Identity-Id"
	headerSessionIdentityType            = "X-Skygear-Session-Identity-Type"
	headerSessionIdentityUpdatedAt       = "X-Skygear-Session-Identity-Updated-At"
	headerSessionAuthenticatorID         = "X-Skygear-Session-Authenticator-Id"
	headerSessionAuthenticatorType       = "X-Skygear-Session-Authenticator-Type"
	headerSessionAuthenticatorOOBChannel = "X-Skygear-Session-Authenticator-Oob-Channel"
	headerSessionAuthenticatorUpdatedAt  = "X-Skygear-Session-Authenticator-Updated-At"
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

	rw.Header().Set(headerSessionIdentityID, i.SessionIdentityID)
	rw.Header().Set(headerSessionIdentityType, string(i.SessionIdentityType))
	rw.Header().Set(headerSessionIdentityUpdatedAt, i.SessionIdentityUpdatedAt.Format(time.RFC3339))

	rw.Header().Set(headerSessionAuthenticatorID, i.SessionAuthenticatorID)
	rw.Header().Set(headerSessionAuthenticatorType, string(i.SessionAuthenticatorType))
	rw.Header().Set(headerSessionAuthenticatorOOBChannel, string(i.SessionAuthenticatorOOBChannel))
	if i.SessionAuthenticatorUpdatedAt != nil {
		rw.Header().Set(headerSessionAuthenticatorUpdatedAt, i.SessionAuthenticatorUpdatedAt.Format(time.RFC3339))
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

	i.SessionIdentityID = r.Header.Get(headerSessionIdentityID)
	i.SessionIdentityType = PrincipalType(r.Header.Get(headerSessionIdentityType))
	if i.SessionIdentityUpdatedAt, err = time.Parse(time.RFC3339, r.Header.Get(headerSessionIdentityUpdatedAt)); err != nil {
		return nil, err
	}

	i.SessionAuthenticatorID = r.Header.Get(headerSessionAuthenticatorID)
	i.SessionAuthenticatorType = AuthenticatorType(r.Header.Get(headerSessionAuthenticatorType))
	i.SessionAuthenticatorOOBChannel = AuthenticatorOOBChannel(r.Header.Get(headerSessionAuthenticatorOOBChannel))
	if updatedAt := r.Header.Get(headerSessionAuthenticatorUpdatedAt); updatedAt != "" {
		updatedAt, err := time.Parse(time.RFC3339, r.Header.Get(headerSessionAuthenticatorUpdatedAt))
		if err != nil {
			return nil, err
		}
		i.SessionAuthenticatorUpdatedAt = &updatedAt
	}

	return i, nil
}
