package facade

import (
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type SessionManager interface {
	List(userID string) ([]session.Session, error)
	Get(id string) (session.Session, error)
}

type SessionFacade struct {
	Sessions SessionManager
}

func (f *SessionFacade) List(userID string) ([]session.Session, error) {
	return f.Sessions.List(userID)
}

func (f *SessionFacade) Get(id string) (session.Session, error) {
	return f.Sessions.Get(id)
}
