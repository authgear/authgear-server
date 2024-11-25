package pgsearch

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var ErrMissingCredential = apierrors.InternalError.WithReason("SearchDisabled").New("search database credential is not provided")
