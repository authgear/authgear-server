package search

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var ErrSearchDisabled = apierrors.InternalError.WithReason("SearchDisabled").New("Search disabled")
