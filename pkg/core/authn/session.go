package authn

type SessionType string

const (
	SessionTypeIdentityProvider SessionType = "idp"
)

type Session interface {
	Attributer

	SessionID() string
	SessionType() SessionType
}
