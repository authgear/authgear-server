package whatsapp

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
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

type WhatsappAPIError struct {
	APIType          config.WhatsappAPIType
	HTTPStatusCode   int
	ResponseBodyText string
	ParsedResponse   *WhatsappAPIErrorResponse
}

var _ error = &WhatsappAPIError{}

func (e *WhatsappAPIError) Error() string {
	return fmt.Sprintf("whatsapp api error: %d", e.HTTPStatusCode)
}

var _ log.LoggingSkippable = &WhatsappAPIError{}

func (e *WhatsappAPIError) SkipLogging() bool {
	switch e.HTTPStatusCode {
	case 401:
		return true
	default:
		if e.ParsedResponse != nil &&
			e.ParsedResponse.Errors != nil &&
			len(*e.ParsedResponse.Errors) > 0 {
			firstErrorCode := (*e.ParsedResponse.Errors)[0].Code
			switch firstErrorCode {
			case errorCodeInvalidUser:
				return true
			}
		}

	}
	return false
}
