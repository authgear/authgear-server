package whatsapp

import (
	"encoding/json"
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
	APIType        config.WhatsappAPIType    `json:"api_type,omitempty"`
	HTTPStatusCode int                       `json:"http_status_code,omitempty"`
	DumpedResponse []byte                    `json:"dumped_response,omitempty"`
	ParsedResponse *WhatsappAPIErrorResponse `json:"-"`
}

var _ error = &WhatsappAPIError{}

func (e *WhatsappAPIError) Error() string {
	jsonText, _ := json.Marshal(e)
	return fmt.Sprintf("whatsapp api error: %s", string(jsonText))
}

var _ log.LoggingSkippable = &WhatsappAPIError{}

func (e *WhatsappAPIError) SkipLogging() bool {
	switch e.HTTPStatusCode {
	case 401:
		return true
	default:
		if e.ParsedResponse != nil {
			if firstErrorCode, ok := e.ParsedResponse.FirstErrorCode(); ok {
				switch firstErrorCode {
				case errorCodeInvalidUser:
					return true
				}
			}
		}

	}
	return false
}
