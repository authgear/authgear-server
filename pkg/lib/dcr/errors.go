package dcr

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrInitialAccessTokenNotFound = apierrors.NotFound.WithReason("InitialAccessTokenNotFound").New("initial access token not found")
