package facade

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/session"
)

type SessionManager interface {
	List(userID string) ([]session.Session, error)
	Get(id string) (session.Session, error)
	RevokeWithEvent(session session.Session, isAdminAPI bool) error
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

	return f.Sessions.RevokeWithEvent(s, true)
}

func (f *SessionFacade) RevokeAll(userID string) error {
	ss, err := f.Sessions.List(userID)
	if err != nil {
		return err
	}

	for _, s := range ss {
		if err := f.Sessions.RevokeWithEvent(s, true); err != nil {
			return err
		}
	}
	return nil
}
