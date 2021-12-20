package handler

type WebSessionType string

const (
	WebSessionTypeCookie       WebSessionType = "cookie"
	WebSessionTypeRefreshToken WebSessionType = "refresh_token"
)

func (p WebSessionType) IsValid() bool {
	switch p {
	case WebSessionTypeCookie:
		return true
	case WebSessionTypeRefreshToken:
		return true
	}
	return false
}
