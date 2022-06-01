package webapp

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var WhatsappWATICallbackSchema = validation.NewMultipartSchema("WhatsappWATICallback")
var _ = WhatsappWATICallbackSchema.Add("WhatsappWATICallback", `
{
	"type": "object",
	"additionalProperties": true,
	"properties": {
		"messages": {
			"type": "array",
			"items": { "$ref": "#/$defs/WhatsappWATICallbackMessage" }
		}
	},
	"required": ["messages"]
}
`)
var _ = WhatsappWATICallbackSchema.Add("WhatsappWATICallbackMessage", `
{
	"type": "object",
	"additionalProperties": true,
	"properties": {
		"from": { "type": "string" },
		"text": {  "$ref": "#/$defs/WhatsappWATICallbackMessageText" }
	},
	"required": ["from"]
}
`)
var _ = WhatsappWATICallbackSchema.Add("WhatsappWATICallbackMessageText", `
{
	"type": "object",
	"additionalProperties": true,
	"properties": {
		"body": { "type": "string" }
	}
}
`)

func init() {
	WhatsappWATICallbackSchema.Instantiate()
}

type WhatsappWATICallbackMessageText struct {
	Body string `json:"body"`
}

type WhatsappWATICallbackMessage struct {
	From string                          `json:"from"`
	Text WhatsappWATICallbackMessageText `json:"text"`
}

type WhatsappWATICallbackRequest struct {
	Messages []WhatsappWATICallbackMessage `json:"messages"`
}

func ConfigureWhatsappWATICallbackRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/whatsapp/callback/wati")
}

type WhatsappWATICallbackHandler struct {
	WhatsappCodeProvider        WhatsappCodeProvider
	Logger                      WhatsappWATICallbackHandlerLogger
	WATICredentials             *config.WATICredentials
	GlobalSessionServiceFactory *GlobalSessionServiceFactory
}

type WhatsappWATICallbackHandlerLogger struct{ *log.Logger }

func NewWhatsappWATICallbackHandlerLogger(lf *log.Factory) WhatsappWATICallbackHandlerLogger {
	return WhatsappWATICallbackHandlerLogger{lf.New("webapp-whatsapp-wati-callback-handler")}
}

func (h *WhatsappWATICallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("wati callback received")

	var err error
	defer func() {
		// always return OK and logs the error if any
		if err != nil {
			h.Logger.WithError(err).Info("failed to consume message")
		}
		w.WriteHeader(http.StatusOK)
	}()

	// check if the auth query matched the one in the secret config
	if h.WATICredentials == nil || h.WATICredentials.WebhookAuth == "" {
		err = errors.New("missing whatsapp.wati secret config")
		return
	}
	authQuery := r.URL.Query().Get("auth")
	if subtle.ConstantTimeCompare([]byte(h.WATICredentials.WebhookAuth), []byte(authQuery)) != 1 {
		err = errors.New("invalid auth query parameters")
		return
	}

	var payload WhatsappWATICallbackRequest
	err = httputil.BindJSONBody(r, w, WhatsappWATICallbackSchema.Validator(), &payload)
	if err != nil {
		return
	}
	if len(payload.Messages) < 1 {
		err = errors.New("missing messages")
		return
	}
	message := payload.Messages[0]
	if message.From == "" {
		err = errors.New("missing message from")
		return
	}
	if message.Text.Body == "" {
		err = errors.New("missing message body")
		return
	}

	phone := message.From
	if !strings.HasPrefix(phone, "+") {
		phone = fmt.Sprintf("+%s", phone)
	}
	code := message.Text.Body
	codeModel, err := h.WhatsappCodeProvider.SetUserInputtedCode(phone, code)
	if err != nil {
		if errors.Is(err, whatsapp.ErrCodeNotFound) {
			err = errors.New("whatsapp code not found")
		}
		return
	}

	// Update the web session and trigger the refresh event
	webSessionProvider := h.GlobalSessionServiceFactory.NewGlobalSessionService(
		config.AppID(codeModel.AppID),
	)
	webSession, err := webSessionProvider.GetSession(codeModel.WebSessionID)
	if err != nil {
		return
	}
	err = webSessionProvider.UpdateSession(webSession)
	if err != nil {
		return
	}
}
