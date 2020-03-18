package authn

type SessionType string

const (
	SessionTypeAuthnInfo SessionType = "authn-info"
)

type Session interface {
	Attributer

	SessionID() string
	SessionType() SessionType
}
