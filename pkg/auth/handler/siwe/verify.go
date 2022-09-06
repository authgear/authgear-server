package siwe

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/api"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
	"github.com/lestrrat-go/jwx/jwk"
	siwego "github.com/spruceid/siwe-go"
)

func ConfigureVerifyRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST").
		WithPathPattern("/siwe/verify")
}

var SIWEVerificationRequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"message": { "type": "string" },
			"signature": { "type": "string" }
		},
		"required": ["message", "signature"]
	}
`)

type VerifyHandlerSIWEService interface {
	VerifyMessage(request apimodel.SIWEVerificationRequest) (*siwego.Message, jwk.Key, error)
}

type VerifyHandlerLogger struct{ *log.Logger }

func NewVerifyHandlerLogger(lf *log.Factory) VerifyHandlerLogger {
	return VerifyHandlerLogger{lf.New("handler-verify")}
}

type VerifyResponse struct {
	Address string  `json:"address"`
	ChainID int     `json:"chain_id"`
	PubKey  jwk.Key `json:"pub_key"`
}

type VerifyHandlerJSONResponseWriter interface {
	WriteResponse(rw http.ResponseWriter, resp *api.Response)
}

type VerifyHandler struct {
	Logger VerifyHandlerLogger
	SIWE   VerifyHandlerSIWEService
	JSON   VerifyHandlerJSONResponseWriter
}

func (h *VerifyHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var payload apimodel.SIWEVerificationRequest
	if err := httputil.BindJSONBody(r, rw, SIWEVerificationRequestSchema.Validator(), &payload); err != nil {
		h.Logger.WithError(err).Error("failed to parse request body")
		http.Error(rw, "bad request", 400)
		return
	}

	msg, pubKey, err := h.SIWE.VerifyMessage(payload)
	if err != nil {
		h.Logger.WithError(err).Error("failed to verify siwe message")
		http.Error(rw, "internal server error", 500)
		return
	}

	address := msg.GetAddress().Hex()
	chainID := msg.GetChainID()

	h.JSON.WriteResponse(rw, &api.Response{
		Result: &VerifyResponse{
			Address: address,
			ChainID: chainID,
			PubKey:  pubKey,
		},
	})
}
