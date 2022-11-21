package facade

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/session"
)

type SessionManager interface {
	List(userID string) ([]session.Session, error)
	Get(id string) (session.Session, error)
	RevokeWithEvent(session session.Session, isTermination bool, isAdminAPI bool) error
	TerminateAllExcept(userID string, idpSessionID string, isAdminAPI bool) error
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

func (f *SessionFacade) Revoke(id string) error {
	s, err := f.Sessions.Get(id)
	if errors.Is(err, session.ErrSessionNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	return f.Sessions.RevokeWithEvent(s, true, true)
}

func (f *SessionFacade) RevokeAll(userID string) error {
	err := f.Sessions.TerminateAllExcept(userID, "", true)
	if err != nil {
		return err
	}
	return nil
}
