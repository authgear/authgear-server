package whatsapp

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrInvalidWhatsappUser = apierrors.BadRequest.
	WithReason("InvalidWhatsappUser").
	New("invalid whatsapp user")
var ErrNoAvailableWhatsappClient = apierrors.BadRequest.
	WithReason("NoAvailableWhatsappClient").
	New("no available whatsapp client")

var ErrUnauthorized = errors.New("whatsapp: unauthorized")
var ErrBadRequest = errors.New("whatsapp: bad request")
var ErrUnexpectedLoginResponse = errors.New("whatsapp: unexpected login response body")
