package resourcescope

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrResourceNotFound = apierrors.NotFound.WithReason("ResourceNotFound").New("resource not found")
var ErrResourceDuplicateURI = apierrors.BadRequest.WithReason("ResourceDuplicateURI").New("duplicate resource uri")
