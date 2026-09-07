package oauthclient

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrDynamicClientNotFound = apierrors.NotFound.WithReason("DynamicClientNotFound").New("dynamic client not found")
var ErrDynamicClientDuplicateClientID = apierrors.BadRequest.WithReason("DynamicClientDuplicateClientID").New("duplicate dynamic client id")

// ErrDynamicClientSourceConflict is an internal invariant violation, not a
// client-visible outcome: it means a client_id that already belongs to
// another source (e.g. a dcrc_-prefixed id colliding with a CIMD URL, which
// cannot happen in practice) was presented to UpsertCIMDClient. See
// Store.UpsertCIMDClient's WHERE guard.
var ErrDynamicClientSourceConflict = apierrors.BadRequest.WithReason("DynamicClientSourceConflict").New("dynamic client id belongs to another source")
