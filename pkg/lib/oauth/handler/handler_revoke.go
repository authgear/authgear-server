package handler

import (
	"crypto/subtle"
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

type RevokeHandler struct {
	OfflineGrants oauth.OfflineGrantStore
	AccessGrants  oauth.AccessGrantStore
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
	if subtle.ConstantTimeCompare([]byte(tokenHash), []byte(offlineGrant.TokenHash)) != 1 {
		return nil
	}

	err = h.OfflineGrants.DeleteOfflineGrant(offlineGrant)
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
