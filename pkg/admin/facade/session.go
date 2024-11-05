package facade

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/session"
)

type SessionManager interface {
	List(ctx context.Context, userID string) ([]session.ListableSession, error)
	Get(ctx context.Context, id string) (session.ListableSession, error)
	RevokeWithEvent(ctx context.Context, session session.SessionBase, isTermination bool, isAdminAPI bool) error
	TerminateAllExcept(ctx context.Context, userID string, currentSession session.ResolvedSession, isAdminAPI bool) error
}

type SessionFacade struct {
	Sessions SessionManager
}

func (f *SessionFacade) List(ctx context.Context, userID string) ([]session.ListableSession, error) {
	return f.Sessions.List(ctx, userID)
}

func (f *SessionFacade) Get(ctx context.Context, id string) (session.ListableSession, error) {
	return f.Sessions.Get(ctx, id)
}

func (f *SessionFacade) Revoke(ctx context.Context, id string) error {
	s, err := f.Sessions.Get(ctx, id)
	if errors.Is(err, session.ErrSessionNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	return f.Sessions.RevokeWithEvent(ctx, s, true, true)
}

func (f *SessionFacade) RevokeAll(ctx context.Context, userID string) error {
	err := f.Sessions.TerminateAllExcept(ctx, userID, nil, true)
	if err != nil {
		return err
	}
	return nil
}
