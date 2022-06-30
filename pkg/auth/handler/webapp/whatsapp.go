package webapp

import (
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	coreimage "github.com/authgear/authgear-server/pkg/util/image"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/boombuler/barcode/qr"
)

const WhatsappMessageOTPPrefix = "#"

var WhatsappMessageOTPRegex = regexp.MustCompile(`#(\d{6})`)

const WhatsappOTPPageQueryXDeviceTokenKey = "x_device_token"
const WhatsappOTPPageQueryStateKey = "state"

type WhatsappOTPPageQueryState string

const (
	WhatsappOTPPageQueryStateInitial     WhatsappOTPPageQueryState = ""
	WhatsappOTPPageQueryStateNoCode      WhatsappOTPPageQueryState = "no_code"
	WhatsappOTPPageQueryStateInvalidCode WhatsappOTPPageQueryState = "invalid_code"
	WhatsappOTPPageQueryStateMatched     WhatsappOTPPageQueryState = "matched"
)

func (s *WhatsappOTPPageQueryState) IsValid() bool {
	return *s == WhatsappOTPPageQueryStateInitial ||
		*s == WhatsappOTPPageQueryStateNoCode ||
		*s == WhatsappOTPPageQueryStateInvalidCode ||
		*s == WhatsappOTPPageQueryStateMatched
}

func getStateFromQuery(r *http.Request) WhatsappOTPPageQueryState {
	p := WhatsappOTPPageQueryState(
		r.URL.Query().Get(WhatsappOTPPageQueryStateKey),
	)
	if p.IsValid() {
		return p
	}
	return WhatsappOTPPageQueryStateInitial
}

type WhatsappCodeProvider interface {
	GetServerWhatsappPhone() string
	VerifyCode(phone string, webSessionID string, consume bool) (*whatsapp.Code, error)
	SetUserInputtedCode(phone string, userInputtedCode string) (*whatsapp.Code, error)
}

type WhatsappOTPViewModel struct {
	PhoneOTPMode               config.AuthenticatorPhoneOTPMode
	WhatsappOTPPhone           string
	WhatsappOTP                string
	UserPhone                  string
	StateQuery                 WhatsappOTPPageQueryState
	OpenWhatsappLink           string // Link to open whatsapp with phone number
	WhatsappCustomURLScheme    htmltemplate.URL
	FormActionPath             string
	OpenWhatsappQRCodeImageURI htmltemplate.URL
	XDeviceToken               bool // value of x_device_token query is used to preserve the checkbox value between whatsapp otp pages
}

func (m *WhatsappOTPViewModel) AddData(r *http.Request, graph *interaction.Graph, codeProvider WhatsappCodeProvider, translations viewmodels.TranslationService) error {
	m.StateQuery = getStateFromQuery(r)
	var n WhatsappOTPNode
	if graph.FindLastNode(&n) {
		m.PhoneOTPMode = n.GetPhoneOTPMode()
		m.WhatsappOTP = n.GetWhatsappOTP()
		m.WhatsappOTPPhone = codeProvider.GetServerWhatsappPhone()
		m.UserPhone = phone.Mask(n.GetPhone())
	}

	q := r.URL.Query()
	// verify code form has auto redirect mechanism
	// clear the state to avoid infinite redirect
	q.Del(WhatsappOTPPageQueryStateKey)
	// clear the x_device_token query to ensure the value is provided by the form data
	q.Del(WhatsappOTPPageQueryXDeviceTokenKey)
	u := url.URL{}
	u.Path = r.URL.Path
	u.RawQuery = q.Encode()
	m.FormActionPath = u.String()

	preFilledMessage, err := translations.RenderText(
		"whatsapp-otp-pre-filled-message",
		map[string]interface{}{
			"target": phone.MaskWithCustomRune(n.GetPhone(), 'x'),
			"otp":    fmt.Sprintf("%s%s", WhatsappMessageOTPPrefix, m.WhatsappOTP),
		},
	)
	if err != nil {
		return err
	}

	waRecipientPhone := strings.TrimPrefix(m.WhatsappOTPPhone, "+")

	// whatsapp universal link
	waURL := url.URL{}
	waURL.Scheme = "https"
	waURL.Host = "wa.me"
	waURL.Path = waRecipientPhone
	q = waURL.Query()
	q.Set("text", preFilledMessage)
	waURL.RawQuery = q.Encode()
	openWhatsappLink := waURL.String()

	// whatsapp custom url scheme
	waCustomURL := url.URL{}
	waCustomURL.Scheme = "whatsapp"
	waCustomURL.Host = "send"
	q = waCustomURL.Query()
	q.Set("text", preFilledMessage)
	q.Set("phone", waRecipientPhone)
	waCustomURL.RawQuery = q.Encode()
	whatsappCustomURLScheme := waCustomURL.String()

	m.OpenWhatsappLink = openWhatsappLink
	// WhatsappCustomURLScheme is generated here and not user generated,
	// so it is safe to use htmltemplate.URL with it.
	// nolint:gosec
	m.WhatsappCustomURLScheme = htmltemplate.URL(whatsappCustomURLScheme)

	// Using H error correction level for adding logo
	img, err := createQRCodeImage(openWhatsappLink, 512, 512, qr.H)
	if err != nil {
		return err
	}
	dataURI, err := coreimage.DataURIFromImage(coreimage.CodecPNG, img)
	if err != nil {
		return err
	}
	// dataURI is generated here and not user generated,
	// so it is safe to use htmltemplate.URL with it.
	// nolint:gosec
	m.OpenWhatsappQRCodeImageURI = htmltemplate.URL(dataURI)

	m.XDeviceToken = r.URL.Query().Get(WhatsappOTPPageQueryXDeviceTokenKey) == "true"

	return nil
}
