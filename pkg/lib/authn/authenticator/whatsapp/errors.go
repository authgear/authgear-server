package whatsapp

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var InvalidWhatsappCode = apierrors.Forbidden.WithReason("InvalidWhatsappCode")

var ErrCodeNotFound = InvalidWhatsappCode.NewWithCause("whatsapp code is expired or invalid", apierrors.StringCause("CodeNotFound"))
var ErrInvalidCode = InvalidWhatsappCode.NewWithCause("invalid whatsapp code", apierrors.StringCause("InvalidWhatsappCode"))
var ErrInputRequired = InvalidWhatsappCode.NewWithCause("whatsapp code not yet received", apierrors.StringCause("InputRequired"))
var ErrWebSessionIDMismatch = InvalidWhatsappCode.NewWithCause("code web session id doesn't match current web session id", apierrors.StringCause("WebSessionIDMismatch"))
