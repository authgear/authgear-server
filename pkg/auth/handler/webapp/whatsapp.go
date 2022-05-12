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
)

const WhatsappOTPPageQueryMethodKey = "method"

type WhatsappOTPPageQueryMethod string

const (
	WhatsappOTPPageQueryMethodInitial      WhatsappOTPPageQueryMethod = ""
	WhatsappOTPPageQueryMethodOpenWhatsapp WhatsappOTPPageQueryMethod = "open_whatsapp"
	WhatsappOTPPageQueryMethodQRCode       WhatsappOTPPageQueryMethod = "qr_code"
)

func (m *WhatsappOTPPageQueryMethod) IsValid() bool {
	return *m == WhatsappOTPPageQueryMethodInitial ||
		*m == WhatsappOTPPageQueryMethodOpenWhatsapp ||
		*m == WhatsappOTPPageQueryMethodQRCode
}

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

func getMethodFromQuery(r *http.Request) WhatsappOTPPageQueryMethod {
	p := WhatsappOTPPageQueryMethod(
		r.URL.Query().Get(WhatsappOTPPageQueryMethodKey),
	)
	if p.IsValid() {
		return p
	}
	return WhatsappOTPPageQueryMethodInitial
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
	VerifyCode(phone string, webSessionID string, consume bool) (*whatsapp.Code, error)
}

type WhatsappOTPViewModel struct {
	PhoneOTPMode               config.AuthenticatorPhoneOTPMode
	WhatsappOTPPhone           string
	WhatsappOTP                string
	MethodQuery                WhatsappOTPPageQueryMethod
	StateQuery                 WhatsappOTPPageQueryState
	OpenWhatsappLink           string // Link to open whatsapp with phone number
	OpenWhatsappPath           string // Path of Auth UI to show open whatsapp link
	ShowQRCodePath             string // Path of Auth UI to open auth ui page to show qr code
	TryAgainPath               string // Path of Auth UI to current method which the state is reset
	OpenWhatsappQRCodeImageURI htmltemplate.URL
}

func (m *WhatsappOTPViewModel) AddData(r *http.Request, graph *interaction.Graph) error {
	m.MethodQuery = getMethodFromQuery(r)
	m.StateQuery = getStateFromQuery(r)
	var n WhatsappOTPNode
	if graph.FindLastNode(&n) {
		m.PhoneOTPMode = n.GetPhoneOTPMode()
		m.WhatsappOTP = n.GetWhatsappOTP()
		// fixme(whatsapp): get whatsapp phone number from config
		m.WhatsappOTPPhone = "+85212345678"
	}

	getPath := func(method string) string {
		q := r.URL.Query()
		q.Set(WhatsappOTPPageQueryMethodKey, method)
		q.Del(WhatsappOTPPageQueryStateKey)
		u := url.URL{}
		u.Path = r.URL.Path
		u.RawQuery = q.Encode()
		return u.String()
	}

	m.OpenWhatsappPath = getPath(string(WhatsappOTPPageQueryMethodOpenWhatsapp))
	m.ShowQRCodePath = getPath(string(WhatsappOTPPageQueryMethodQRCode))
	currentMethod := getMethodFromQuery(r)
	m.TryAgainPath = getPath(string(currentMethod))

	waURL := url.URL{}
	waURL.Scheme = "https"
	waURL.Host = "wa.me"
	waURL.Path = strings.TrimPrefix(m.WhatsappOTPPhone, "+")
	q := waURL.Query()
	q.Set("text", m.WhatsappOTP)
	waURL.RawQuery = q.Encode()
	m.OpenWhatsappLink = waURL.String()

	img, err := createQRCodeImage(m.OpenWhatsappLink, 512, 512)
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
