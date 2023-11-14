package web

import "net/http"

const LoginLinkOTPPageQueryStateKey = "state"

type LoginLinkOTPPageQueryState string

const (
	LoginLinkOTPPageQueryStateInitial     LoginLinkOTPPageQueryState = ""
	LoginLinkOTPPageQueryStateInvalidCode LoginLinkOTPPageQueryState = "invalid_code"
	LoginLinkOTPPageQueryStateMatched     LoginLinkOTPPageQueryState = "matched"
)

func (s *LoginLinkOTPPageQueryState) IsValid() bool {
	return *s == LoginLinkOTPPageQueryStateInitial ||
		*s == LoginLinkOTPPageQueryStateInvalidCode ||
		*s == LoginLinkOTPPageQueryStateMatched
}

func GetLoginLinkStateFromQuery(r *http.Request) LoginLinkOTPPageQueryState {
	p := LoginLinkOTPPageQueryState(
		r.URL.Query().Get(LoginLinkOTPPageQueryStateKey),
	)
	if p.IsValid() {
		return p
	}
	return LoginLinkOTPPageQueryStateInitial
}
