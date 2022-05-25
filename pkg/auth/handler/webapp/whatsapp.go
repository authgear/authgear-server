package webapp

import (
	htmltemplate "html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	coreimage "github.com/authgear/authgear-server/pkg/util/image"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/boombuler/barcode/qr"
)

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
	VerifyFormSubmitPath       string // Path of Auth UI for verify form submission which the state is reset
	OpenWhatsappQRCodeImageURI htmltemplate.URL
}

func (m *WhatsappOTPViewModel) AddData(r *http.Request, graph *interaction.Graph, codeProvider WhatsappCodeProvider) error {
	m.StateQuery = getStateFromQuery(r)
	var n WhatsappOTPNode
	if graph.FindLastNode(&n) {
		m.PhoneOTPMode = n.GetPhoneOTPMode()
		m.WhatsappOTP = n.GetWhatsappOTP()
		m.WhatsappOTPPhone = codeProvider.GetServerWhatsappPhone()
		m.UserPhone = phone.Mask(n.GetPhone())
	}

	getPath := func() string {
		q := r.URL.Query()
		// delete the state in query is intended
		q.Del(WhatsappOTPPageQueryStateKey)
		u := url.URL{}
		u.Path = r.URL.Path
		u.RawQuery = q.Encode()
		return u.String()
	}

	// verify code form has auto redirect mechanism
	// reset the state of VerifyFormSubmitPath to avoid infinite redirect
	m.VerifyFormSubmitPath = getPath()

	waURL := url.URL{}
	waURL.Scheme = "https"
	waURL.Host = "wa.me"
	waURL.Path = strings.TrimPrefix(m.WhatsappOTPPhone, "+")
	q := waURL.Query()
	q.Set("text", m.WhatsappOTP)
	waURL.RawQuery = q.Encode()
	m.OpenWhatsappLink = waURL.String()

	// Using H error correction level for adding logo
	img, err := createQRCodeImage(m.OpenWhatsappLink, 512, 512, qr.H)
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

	return nil
}
