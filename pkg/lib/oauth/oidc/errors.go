package oidc

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrInvalidCustomURI = apierrors.Invalid.WithReason("WebUIInvalidCustomURI")
