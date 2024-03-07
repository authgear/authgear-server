package oidc

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrInvalidCustomURI = apierrors.Invalid.WithReason("WebUIInvalidCustomURI")
var ErrInvalidSettingsAction = apierrors.Invalid.WithReason("WebUIInvalidSettingsAction")
