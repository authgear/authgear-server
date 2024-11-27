package webapp

import (
	"context"
	"net/http"

	"net/url"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsProfileEditHTML = template.RegisterHTML(
	"web/settings_profile_edit.html",
	Components...,
)

func ConfigureSettingsProfileEditRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/settings/profile/:variant/edit")
}

type SettingsProfileEditUserService interface {
	GetRaw(ctx context.Context, id string) (*user.User, error)
}

type SettingsProfileEditStdAttrsService interface {
	UpdateStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type SettingsProfileEditCustomAttrsService interface {
	UpdateCustomAttributesWithForm(ctx context.Context, role accesscontrol.Role, userID string, jsonPointerMap map[string]string) error
}

type SettingsProfileEditHandler struct {
	ControllerFactory        ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 Renderer
	Users                    SettingsProfileEditUserService
	StdAttrs                 SettingsProfileEditStdAttrsService
	CustomAttrs              SettingsProfileEditCustomAttrsService
}

func (h *SettingsProfileEditHandler) GetData(ctx context.Context, r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
	userID := session.GetUserID(ctx)

	data := map[string]interface{}{}

	variant := httproute.GetParam(r, "variant")
	data["Variant"] = variant
	data["Pointer"] = r.FormValue("pointer")

	baseViewModel := h.BaseViewModel.ViewModel(r, rw)
	viewmodels.Embed(data, baseViewModel)

	viewModelPtr, err := h.SettingsProfileViewModel.ViewModel(ctx, *userID)
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
	defer ctrl.ServeWithDBTx(r.Context())

	ctrl.Get(func(ctx context.Context) error {
		data, err := h.GetData(ctx, r, w)
		if err != nil {
			return err
		}

		h.Renderer.RenderHTML(w, r, TemplateWebSettingsProfileEditHTML, data)
		return nil
	})

	ctrl.PostAction("save", func(ctx context.Context) error {
		userID := *session.GetUserID(ctx)
		PatchGenderForm(r.Form)
		m := JSONPointerFormToMap(r.Form)

		u, err := h.Users.GetRaw(ctx, userID)
		if err != nil {
			return err
		}

		variant := httproute.GetParam(r, "variant")
		if variant == "custom_attributes" {
			err = h.CustomAttrs.UpdateCustomAttributesWithForm(ctx, config.RoleEndUser, userID, m)
			if err != nil {
				return err
			}
		} else {
			attrs, err := stdattrs.T(u.StandardAttributes).MergedWithForm(m)
			if err != nil {
				return err
			}

			err = h.StdAttrs.UpdateStandardAttributes(ctx, config.RoleEndUser, userID, attrs)
			if err != nil {
				return err
			}
		}

		result := webapp.Result{RedirectURI: "/settings/profile"}
		result.WriteResponse(w, r)
		return nil
	})
}

func PatchGenderForm(form url.Values) {
	_, genderSelectOK := form["gender-select"]
	if !genderSelectOK {
		return
	}

	genderSelect := form.Get("gender-select")
	genderInput := form.Get("gender-input")

	if genderSelect == "other" {
		form.Set("/gender", genderInput)
	} else {
		form.Set("/gender", genderSelect)
	}
}
