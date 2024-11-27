package authflowv2

import (
	"context"
	htmltemplate "html/template"
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"

	"github.com/authgear/authgear-server/pkg/lib/accountmanagement"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	coreimage "github.com/authgear/authgear-server/pkg/util/image"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/template"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var TemplateWebSettingsMFACreateTOTPHTML = template.RegisterHTML(
	"web/authflowv2/settings_mfa_create_totp.html",
	handlerwebapp.SettingsComponents...,
)

var AuthflowV2SettingsMFACreateTOTPSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"properties": {
			"x_totp": { "type": "string" },
			"x_confirm_totp": { "type": "string" }
		},
		"required": ["x_totp", "x_confirm_totp"]
	}
`)

func ConfigureAuthflowV2SettingsMFACreateTOTPRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "POST", "GET").
		WithPathPattern(AuthflowV2RouteSettingsMFACreateTOTP)
}

type AuthflowV2SettingsMFACreateTOTPViewModel struct {
	Token    string
	Secret   string
	ImageURI htmltemplate.URL
}

type AuthflowV2SettingsMFACreateTOTPHandler struct {
	Database          *appdb.Handle
	ControllerFactory handlerwebapp.ControllerFactory
	BaseViewModel     *viewmodels.BaseViewModeler
	SettingsViewModel *viewmodels.SettingsViewModeler
	Renderer          handlerwebapp.Renderer

	AccountManagement *accountmanagement.Service
}

func (h *AuthflowV2SettingsMFACreateTOTPHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter, tokenString string, totpSecret string, otpauthURI string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	userID := session.GetUserID(ctx)

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	img, err := secretcode.QRCodeImageFromURI(otpauthURI, 512, 512)
	if err != nil {
		return nil, err
	}
	dataURI, err := coreimage.DataURIFromImage(coreimage.CodecPNG, img)
	if err != nil {
		return nil, err
	}

	// SettingsViewModel
	viewModelPtr, err := h.SettingsViewModel.ViewModel(ctx, *userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *viewModelPtr)

	screenViewModel := AuthflowV2SettingsMFACreateTOTPViewModel{
		Token:  tokenString,
		Secret: totpSecret,
		// nolint: gosec
		ImageURI: htmltemplate.URL(dataURI),
	}
	viewmodels.Embed(data, screenViewModel)

	return data, nil
}

func (h *AuthflowV2SettingsMFACreateTOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.ServeWithoutDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		s := session.GetSession(ctx)

		tokenString := r.Form.Get("q_token")
		token, err := h.AccountManagement.GetToken(ctx, s, tokenString)
		if err != nil {
			return err
		}

		opts := secretcode.URIOptions{
			Issuer:      string(token.Authenticator.TOTPIssuer),
			AccountName: token.Authenticator.TOTPEndUserAccountID,
		}
		totp, err := secretcode.NewTOTPFromSecret(token.Authenticator.TOTPSecret)
		if err != nil {
			return err
		}
		totpauthURI := totp.GetURI(opts).String()

		var data map[string]interface{}
		err = h.Database.WithTx(ctx, func(ctx context.Context) error {
			data, err = h.GetData(ctx, r, w, tokenString, totp.Secret, totpauthURI)
			return err
		})
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsMFACreateTOTPHTML, data)
		return nil
	})
}
