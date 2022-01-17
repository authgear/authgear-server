package handler

type WebSessionType string

const (
	WebSessionTypeCookie       WebSessionType = "cookie"
	WebSessionTypeRefreshToken WebSessionType = "refresh_token"
)
