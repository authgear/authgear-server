package handler

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

type SessionManager interface {
	RevokeWithEvent(ctx context.Context, session session.SessionBase, isTermination bool, isAdminAPI bool) error
	RevokeWithoutEvent(ctx context.Context, session session.SessionBase) error
}

type RevokeHandlerOfflineGrantService interface {
	GetOfflineGrant(ctx context.Context, id string) (*oauth.OfflineGrant, error)
}

type RevokeHandlerAccessGrantStore interface {
	GetAccessGrant(ctx context.Context, tokenHash string) (*oauth.AccessGrant, error)
	DeleteAccessGrant(ctx context.Context, g *oauth.AccessGrant) error
}

type RevokeHandler struct {
	SessionManager      SessionManager
	OfflineGrantService RevokeHandlerOfflineGrantService
	AccessGrants        RevokeHandlerAccessGrantStore
}

func (h *RevokeHandler) Handle(ctx context.Context, r protocol.RevokeRequest) error {
	token, grantID, err := oauth.DecodeRefreshToken(r.Token())
	if err == nil {
		return h.revokeOfflineGrant(ctx, token, grantID)
	}
	return h.revokeAccessGrant(ctx, r.Token())
}

func (h *RevokeHandler) revokeOfflineGrant(ctx context.Context, token, grantID string) error {
	offlineGrant, err := h.OfflineGrantService.GetOfflineGrant(ctx, grantID)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	tokenHash := oauth.HashToken(token)
	if !offlineGrant.MatchCurrentHash(tokenHash) {
		return nil
	}

	err = h.SessionManager.RevokeWithEvent(ctx, offlineGrant, false, false)
	if err != nil {
		return err
	}

	return nil
}

func (h *RevokeHandler) revokeAccessGrant(ctx context.Context, token string) error {
	tokenHash := oauth.HashToken(token)
	accessGrant, err := h.AccessGrants.GetAccessGrant(ctx, tokenHash)
	if errors.Is(err, oauth.ErrGrantNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	err = h.AccessGrants.DeleteAccessGrant(ctx, accessGrant)
	if err != nil {
		return err
	}

	return nil
}
