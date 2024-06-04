package webappoauth

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrOAuthStateInvalid = apierrors.NewInvalid("invalid state")
