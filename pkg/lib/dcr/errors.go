package dcr

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrInitialAccessTokenNotFound = apierrors.NotFound.WithReason("InitialAccessTokenNotFound").New("initial access token not found")

// ErrInitialAccessTokenExpired is distinct from ErrInitialAccessTokenNotFound
// for the audit log only -- the HTTP response maps both to the identical
// invalid_initial_access_token error, so an unauthenticated caller learns
// nothing from which one actually happened. The project admin's audit
// record is a different, authenticated channel: "expired" is a broken
// integration to reissue, "unknown" is guessing or a revoked token still in
// use, and conflating them there would make the record nearly useless.
var ErrInitialAccessTokenExpired = apierrors.NotFound.WithReason("InitialAccessTokenExpired").New("initial access token expired")
