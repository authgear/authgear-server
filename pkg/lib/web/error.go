package web

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var UnsupportedImageFile = apierrors.BadRequest.WithReason("UnsupportedImageFile")
