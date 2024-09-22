package userexport

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var ErrUserExportDisabled = apierrors.InternalError.WithReason("UserExportDisabled").New("User export disabled")
