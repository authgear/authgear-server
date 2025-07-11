package resourcescope

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrResourceNotFound = apierrors.NotFound.WithReason("ResourceNotFound").New("resource not found")
var ErrResourceDuplicateURI = apierrors.BadRequest.WithReason("ResourceDuplicateURI").New("duplicate resource uri")

var ErrScopeNotFound = apierrors.NotFound.WithReason("ScopeNotFound").New("scope not found")
var ErrScopeDuplicate = apierrors.BadRequest.WithReason("ScopeDuplicate").New("duplicate scope")

var ErrClientNotFound = apierrors.NotFound.WithReason("ClientNotFound").New("client not found")
