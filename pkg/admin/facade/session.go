package facade

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/session"
)

type SessionManager interface {
	List(userID string) ([]session.ListableSession, error)
	Get(id string) (session.ListableSession, error)
	RevokeWithEvent(session session.SessionBase, isTermination bool, isAdminAPI bool) error
	TerminateAllExcept(userID string, currentSession session.ResolvedSession, isAdminAPI bool) error
}

type SessionFacade struct {
	Sessions SessionManager
}

func (f *SessionFacade) List(userID string) ([]session.ListableSession, error) {
	return f.Sessions.List(userID)
}

func (f *SessionFacade) Get(id string) (session.ListableSession, error) {
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
	err := f.Sessions.TerminateAllExcept(userID, nil, true)
	if err != nil {
		return err
	}
	return nil
}
