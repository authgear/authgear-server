package oidc

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
)

var ErrInvalidCustomURI = apierrors.Invalid.WithReason("WebUIInvalidCustomURI")

func NewErrInvalidSettingsAction(errMsg string) error {
	return protocol.NewError("invalid_request", errMsg)
}
