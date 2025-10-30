package smsapi

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/translation"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

var NoAvailableClient = apierrors.InternalError.
	WithReason("NoAvailableSMSClient")
var ErrNoAvailableClient = NoAvailableClient.
	New("no available SMS client")
var ErrAmbiguousClient = errors.New("ambiguous SMS client")

var ErrKindInvalidPhoneNumber = apierrors.BadRequest.WithReason("SMSGatewayInvalidPhoneNumber")
var ErrKindAuthenticationFailed = apierrors.InternalError.WithReason("SMSGatewayAuthenticationFailed")
var ErrKindDeliveryRejected = apierrors.InternalError.WithReason("SMSGatewayDeliveryRejected")
var ErrKindTimeout = apierrors.InternalError.WithReason("SMSGatewayTimeout")
var ErrKindRateLimited = apierrors.TooManyRequest.WithReason("SMSGatewayRateLimited")
var ErrKindUnsupportedRequest = apierrors.TooManyRequest.WithReason("SMSGatewayUnsupportedRequest")

type TemplateVariables struct {
	AppName     string `json:"app_name,omitempty"`
	ClientID    string `json:"client_id,omitempty"`
	Code        string `json:"code,omitempty"`
	Email       string `json:"email,omitempty"`
	HasPassword bool   `json:"has_password,omitempty"`
	Host        string `json:"host,omitempty"`
	Link        string `json:"link,omitempty"`
	Password    string `json:"password,omitempty"`
	Phone       string `json:"phone,omitempty"`
	State       string `json:"state,omitempty"`
	UILocales   string `json:"ui_locales,omitempty"`
	URL         string `json:"url,omitempty"`
	XState      string `json:"x_state,omitempty"`
}

func NewTemplateVariablesFromPreparedTemplateVariables(v *translation.PreparedTemplateVariables) *TemplateVariables {
	return &TemplateVariables{
		AppName:     v.AppName,
		ClientID:    v.ClientID,
		Code:        v.Code,
		Email:       v.Email,
		HasPassword: v.HasPassword,
		Host:        v.Host,
		Link:        v.Link,
		Password:    v.Password,
		Phone:       v.Phone,
		State:       v.State,
		UILocales:   v.UILocales,
		URL:         v.URL,
		XState:      v.XState,
	}
}

type SendOptions struct {
	Sender            string
	To                string
	Body              string
	AppID             string
	TemplateName      string
	LanguageTag       string
	TemplateVariables *TemplateVariables
}

type Client interface {
	Send(ctx context.Context, opts SendOptions) error
}

type SendError struct {
	DumpedResponse []byte             `json:"dumped_response,omitempty"`
	APIErrorKind   *apierrors.Kind    `json:"api_error_kind,omitempty"`
	ProviderType   config.SMSProvider `json:"provider_type,omitempty"`

	ProviderErrorCode string `json:"provider_error_code,omitempty"`

	// provider_type=custom
	CustomProviderName         string `json:"custom_provider_name,omitempty"`
	CustomProviderResponseCode string `json:"custom_provider_response_code,omitempty"`
	CustomProviderDescription  string `json:"description,omitempty"`
}

func (e *SendError) Error() string {
	jsonText, _ := json.Marshal(e)
	return string(jsonText)
}

func (e *SendError) As(target any) bool {
	switch target.(type) {
	case **apierrors.APIError:
		apierr := e.asAPIError()
		*target.(**apierrors.APIError) = apierr
		return true
	case *errorutil.Detailer:
		apierr := e.asAPIError()
		*target.(*errorutil.Detailer) = apierr
		return true
	}
	return false
}

func (e *SendError) asAPIError() *apierrors.APIError {
	details := apierrors.Details{
		"ProviderErrorCode": e.ProviderErrorCode,
		"ProviderType":      e.ProviderType,
		"ProviderName":      e.CustomProviderName,
		"Description":       e.CustomProviderDescription,
	}
	kind := apierrors.UnexpectedError
	if e.APIErrorKind != nil {
		kind = *e.APIErrorKind
	}
	return kind.NewWithInfo("sms gateway send error", details)
}
