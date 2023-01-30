package webapp

import "net/http"

const MagicLinkOTPPageQueryStateKey = "state"

type MagicLinkOTPPageQueryState string

const (
	MagicLinkOTPPageQueryStateInitial     MagicLinkOTPPageQueryState = ""
	MagicLinkOTPPageQueryStateInvalidCode MagicLinkOTPPageQueryState = "invalid_code"
	MagicLinkOTPPageQueryStateMatched     MagicLinkOTPPageQueryState = "matched"
)

func (s *MagicLinkOTPPageQueryState) IsValid() bool {
	return *s == MagicLinkOTPPageQueryStateInitial ||
		*s == MagicLinkOTPPageQueryStateInvalidCode ||
		*s == MagicLinkOTPPageQueryStateMatched
}

func GetMagicLinkStateFromQuery(r *http.Request) MagicLinkOTPPageQueryState {
	p := MagicLinkOTPPageQueryState(
		r.URL.Query().Get(MagicLinkOTPPageQueryStateKey),
	)
	if p.IsValid() {
		return p
	}
	return MagicLinkOTPPageQueryStateInitial
}
