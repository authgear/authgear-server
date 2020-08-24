package model

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type SessionInfo struct {
	IsValid       bool
	UserID        string
	UserAnonymous bool
	UserVerified  bool

	SessionACR string
	SessionAMR []string
}

const (
	headerSessionValid  = "X-Authgear-Session-Valid"
	headerUserID        = "X-Authgear-User-Id"
	headerUserVerified  = "X-Authgear-User-Verified"
	headerUserAnonymous = "X-Authgear-User-Anonymous"
	headerSessionAcr    = "X-Authgear-Session-Acr"
	headerSessionAmr    = "X-Authgear-Session-Amr"
)

func (i *SessionInfo) PopulateHeaders(rw http.ResponseWriter) {
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

func headerParseBool(name string, value string) (b bool, err error) {
	b, err = strconv.ParseBool(value)
	if err != nil {
		err = fmt.Errorf("session: failed to parse %v: %w", name, err)
	}
	return
}

func headerParseSpaceSeparated(value string) (ss []string) {
	if value == "" {
		return
	}
	ss = strings.Split(value, " ")
	return
}

func NewSessionInfoFromHeaders(hdr http.Header) (info *SessionInfo, err error) {
	sessionValidStr := hdr.Get(headerSessionValid)
	if sessionValidStr == "" {
		return nil, nil
	}

	info = &SessionInfo{}
	sessionValid, err := headerParseBool(headerSessionValid, sessionValidStr)
	if err != nil {
		return
	}
	if !sessionValid {
		return
	}

	userID := hdr.Get(headerUserID)

	anonymous, err := headerParseBool(headerUserAnonymous, hdr.Get(headerUserAnonymous))
	if err != nil {
		return
	}

	verified, err := headerParseBool(headerUserVerified, hdr.Get(headerUserVerified))
	if err != nil {
		return
	}

	acr := hdr.Get(headerSessionAcr)
	if err != nil {
		return
	}

	amr := headerParseSpaceSeparated(hdr.Get(headerSessionAmr))

	info.IsValid = sessionValid
	info.UserID = userID
	info.UserAnonymous = anonymous
	info.UserVerified = verified
	info.SessionACR = acr
	info.SessionAMR = amr
	return
}
