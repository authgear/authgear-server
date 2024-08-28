package authflowv2

import (
	"net/http"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func init() {
	settingsProfileEditVariantToTemplate = make(map[string]*template.HTML)
}

var settingsProfileEditVariantToTemplate map[string]*template.HTML

type SettingsProfileEditHandler struct {
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 handlerwebapp.Renderer

	Users       handlerwebapp.SettingsProfileEditUserService
	StdAttrs    handlerwebapp.SettingsProfileEditStdAttrsService
	CustomAttrs handlerwebapp.SettingsProfileEditCustomAttrsService
}

func (h *SettingsProfileEditHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	userID := session.GetUserID(r.Context())

	data := map[string]interface{}{}

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	viewModelPtr, err := h.SettingsProfileViewModel.ViewModel(*userID)
	if err != nil {
		return nil, err
	}
	viewmodels.Embed(data, *viewModelPtr)

	return data, nil
}

func (h *SettingsProfileEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl, err := h.ControllerFactory.New(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer ctrl.Serve()

	ctrl.Get(func() error {
		data, err := h.GetData(r, w)
		if err != nil {
			return err
		}

		variant := httproute.GetParam(r, "variant")
		settingsTemplate, ok := settingsProfileEditVariantToTemplate[variant]
		if !ok {
			h.Renderer.RenderHTML(w, r, TemplateWebNotFoundHTML, data)
			return nil
		}

		h.Renderer.RenderHTML(w, r, settingsTemplate, data)
		return nil
	})

	ctrl.PostAction("save", func() error {
		userID := *session.GetUserID(r.Context())
		m := handlerwebapp.JSONPointerFormToMap(r.Form)

		u, err := h.Users.GetRaw(userID)
		if err != nil {
			return err
		}

		variant := httproute.GetParam(r, "variant")
		if variant == "custom_attributes" {
			err = h.CustomAttrs.UpdateCustomAttributesWithForm(config.RoleEndUser, userID, m)
			if err != nil {
				return err
			}
		} else {
			attrs, err := stdattrs.T(u.StandardAttributes).MergedWithForm(m)
			if err != nil {
				return err
			}

			err = h.StdAttrs.UpdateStandardAttributes(config.RoleEndUser, userID, attrs)
			if err != nil {
				return err
			}
		}

		result := webapp.Result{RedirectURI: "/settings/profile"}
		result.WriteResponse(w, r)
		return nil
	})
}
