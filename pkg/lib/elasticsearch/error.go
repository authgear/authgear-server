package elasticsearch

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var ErrMissingCredential = apierrors.InternalError.WithReason("SearchDisabled").New("elasticsearch credential is not provided")
