package mfa

import (
	"bytes"
	"image/png"
	"net/http"
	"net/url"
	"strconv"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"

	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachTOTPQRCodeHandler(
	server *server.Server,
	authDependency auth.DependencyMap,
) *server.Server {
	server.Handle("/mfa/totp/qrcode", &TOTPQRCodeHandlerFactory{
		Dependency: authDependency,
	}).Methods("OPTIONS", "GET")
	return server
}

type TOTPQRCodeHandlerFactory struct {
	Dependency auth.DependencyMap
}

func (f TOTPQRCodeHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &TOTPQRCodeHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	return h
}

/*
	@Operation GET /mfa/totp/qrcode - Generate QR code image
		Generate QR code image of the given key URI.

		@Response 200
			QR code image
*/
type TOTPQRCodeHandler struct{}

func (h *TOTPQRCodeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			handler.WriteResponse(w, handler.APIResponse{Err: skyerr.MakeError(err)})
		}
	}()

	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return
	}

	otpauthURI := q.Get("otpauth_uri")
	keyURI, err := mfa.ParseKeyURI(otpauthURI)
	if err != nil {
		return
	}
	if !keyURI.IsGoogleAuthenticatorCompatible() {
		err = skyerr.NewInvalidArgument("otpauth_uri is not Google Authenticator Compatiable", []string{"otpauth_uri"})
		return
	}

	img, err := qr.Encode(otpauthURI, qr.M, qr.Auto)
	if err != nil {
		return
	}

	img, err = barcode.Scale(img, 512, 512)
	if err != nil {
		return
	}

	buf := &bytes.Buffer{}
	err = png.Encode(buf, img)
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
