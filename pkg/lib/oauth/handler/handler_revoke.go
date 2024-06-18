package handler

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type SessionManager interface {
	RevokeWithEvent(session session.SessionBase, isTermination bool, isAdminAPI bool) error
	RevokeWithoutEvent(session session.SessionBase) error
}

type RevokeHandler struct {
	SessionManager SessionManager
	OfflineGrants  oauth.OfflineGrantStore
	AccessGrants   oauth.AccessGrantStore
}

func (h *RevokeHandler) Handle(r protocol.RevokeRequest) error {
	token, grantID, err := oauth.DecodeRefreshToken(r.Token())
	if err == nil {
		return h.revokeOfflineGrant(token, grantID)
	}
	return h.revokeAccessGrant(r.Token())
}

func (h *RevokeHandler) revokeOfflineGrant(token, grantID string) error {
	offlineGrant, err := h.OfflineGrants.GetOfflineGrant(grantID)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	tokenHash := oauth.HashToken(token)
	if !offlineGrant.MatchHash(tokenHash) {
		return nil
	}

	err = h.SessionManager.RevokeWithEvent(offlineGrant, false, false)
	if err != nil {
		return err
	}

	return nil
}

func (h *RevokeHandler) revokeAccessGrant(token string) error {
	tokenHash := oauth.HashToken(token)
	accessGrant, err := h.AccessGrants.GetAccessGrant(tokenHash)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	err = h.AccessGrants.DeleteAccessGrant(accessGrant)
	if err != nil {
		return err
	}

	return nil
}
