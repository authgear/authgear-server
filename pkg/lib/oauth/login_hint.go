package oauth

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type LoginHintType string

const (
	LoginHintTypeAnonymous LoginHintType = "anonymous"
	// nolint: gosec
	LoginHintTypeAppSessionToken LoginHintType = "app_session_token"
	LoginHintTypeLoginID         LoginHintType = "login_id"
)

type LoginHint struct {
	Type    LoginHintType
	Enforce bool

	// Specific to LoginHintTypeAnonymous
	PromotionCode string
	JWT           string

	// Specific to LoginHintTypeAppSessionToken
	AppSessionToken string

	// Specific to LoginHintTypeLoginID
	LoginIDEmail    string
	LoginIDUsername string
	LoginIDPhone    string
}

func ParseLoginHint(s string) (*LoginHint, error) {
	if !strings.HasPrefix(s, "https://authgear.com/login_hint?") {
		return nil, fmt.Errorf("invalid login_hint: %v", s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("login_hint is not an URL: %w", err)
	}
	q := u.Query()

	var loginHint LoginHint

	typ := q.Get("type")
	enforce, err := strconv.ParseBool(q.Get("enforce"))
	if err != nil {
		enforce = false
	}
	loginHint.Enforce = enforce

	switch typ {
	case string(LoginHintTypeAnonymous):
		loginHint.Type = LoginHintTypeAnonymous
		loginHint.PromotionCode = q.Get("promotion_code")
		loginHint.JWT = q.Get("jwt")
	case string(LoginHintTypeAppSessionToken):
		loginHint.Type = LoginHintTypeAppSessionToken
		loginHint.AppSessionToken = q.Get("app_session_token")
	case string(LoginHintTypeLoginID):
		loginHint.Type = LoginHintTypeLoginID
		loginHint.LoginIDEmail = q.Get("email")
		loginHint.LoginIDPhone = q.Get("phone")
		loginHint.LoginIDUsername = q.Get("username")
	default:
		return nil, fmt.Errorf("invalid login_hint type: %v", typ)
	}

	return &loginHint, nil
}
