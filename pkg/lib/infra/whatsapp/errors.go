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

const (
	onPremisesErrorCodeInvalidUser = 1013
	// The list of error codes.
	// https://developers.facebook.com/docs/whatsapp/cloud-api/support/error-codes/
	// This code could possibly means invalid whatsapp user.
	// However, in my own testing, the response is still successful even if I send to
	// non-whatsapp number.
	cloudAPIErrorCodeMaybeInvalidUser = 131026
)

type WhatsappAPIError struct {
	APIType            config.WhatsappAPIType              `json:"api_type,omitempty"`
	HTTPStatusCode     int                                 `json:"http_status_code,omitempty"`
	DumpedResponse     []byte                              `json:"dumped_response,omitempty"`
	OnPremisesResponse *WhatsappOnPremisesAPIErrorResponse `json:"-"`
	CloudAPIResponse   *WhatsappCloudAPIErrorResponse      `json:"-"`
}

func (e *WhatsappAPIError) GetErrorCode() (int, bool) {
	if e.APIType == config.WhatsappAPITypeOnPremises && e.OnPremisesResponse != nil {
		return e.OnPremisesResponse.FirstErrorCode()
	}
	if e.APIType == config.WhatsappAPITypeCloudAPI && e.CloudAPIResponse != nil {
		return e.CloudAPIResponse.Error.Code, true
	}
	return 0, false
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
		if e.OnPremisesResponse != nil {
			if firstErrorCode, ok := e.OnPremisesResponse.FirstErrorCode(); ok {
				if firstErrorCode == onPremisesErrorCodeInvalidUser {
					return true
				}
			}
		}
		if e.CloudAPIResponse != nil {
			if e.CloudAPIResponse.Error.Code == cloudAPIErrorCodeMaybeInvalidUser {
				return true
			}
		}

	}
	return false
}
