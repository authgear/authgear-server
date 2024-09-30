package userexport

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var ErrUserExportDisabled = apierrors.InternalError.WithReason("UserExportDisabled").New("User export disabled")
var ErrUserExportDuplicateField = apierrors.Invalid.WithReason("UserExportNonUniqueFieldNames")
