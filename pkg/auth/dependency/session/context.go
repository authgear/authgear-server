package session

type Context interface {
	Session() *Session
}
