package authflowv2

import (
	"net/http"
	"net/url"

	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/template"
)

func init() {
	settingsProfileEditVariantToTemplate = make(map[string]*template.HTML)
	settingsProfileEditVariantToTemplate["picture"] = template.RegisterHTML(
		"web/authflowv2/settings_picture_edit.html",
		handlerwebapp.SettingsComponents...,
	)
	settingsProfileEditVariantToTemplate["birthdate"] = template.RegisterHTML(
		"web/authflowv2/settings_profile_edit_birthdate.html",
		handlerwebapp.SettingsComponents...,
	)
	settingsProfileEditVariantToTemplate["gender"] = template.RegisterHTML(
		"web/authflowv2/settings_profile_edit_gender.html",
		handlerwebapp.SettingsComponents...,
	)
	settingsProfileEditVariantToTemplate["locale"] = template.RegisterHTML(
		"web/authflowv2/settings_profile_edit_locale.html",
		handlerwebapp.SettingsComponents...,
	)
	settingsProfileEditVariantToTemplate["name"] = template.RegisterHTML(
		"web/authflowv2/settings_profile_edit_name.html",
		handlerwebapp.SettingsComponents...,
	)
	settingsProfileEditVariantToTemplate["zoneinfo"] = template.RegisterHTML(
		"web/authflowv2/settings_profile_edit_zoneinfo.html",
		handlerwebapp.SettingsComponents...,
	)
}

var settingsProfileEditVariantToTemplate map[string]*template.HTML

var TemplateSettingsProfileNoPermission = template.RegisterHTML(
	"web/authflowv2/settings_profile_no_permission.html",
	handlerwebapp.Components...,
)

type AuthflowV2SettingsProfileEditHandler struct {
	ControllerFactory        handlerwebapp.ControllerFactory
	BaseViewModel            *viewmodels.BaseViewModeler
	SettingsProfileViewModel *viewmodels.SettingsProfileViewModeler
	Renderer                 handlerwebapp.Renderer

	UserProfileConfig *config.UserProfileConfig

	Users       handlerwebapp.SettingsProfileEditUserService
	StdAttrs    handlerwebapp.SettingsProfileEditStdAttrsService
	CustomAttrs handlerwebapp.SettingsProfileEditCustomAttrsService
}

func (h *AuthflowV2SettingsProfileEditHandler) GetData(r *http.Request, rw http.ResponseWriter) (map[string]interface{}, error) {
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

func (h *AuthflowV2SettingsProfileEditHandler) isAttributeEditable(attributeVariant string) bool {
	accessControl := h.UserProfileConfig.StandardAttributes.GetAccessControl().MergedWith(
		h.UserProfileConfig.CustomAttributes.GetAccessControl(),
	)

	isEditable := func(jsonpointer string) bool {
		level := accessControl.GetLevel(
			accesscontrol.Subject(jsonpointer),
			config.RoleEndUser,
			config.AccessControlLevelHidden,
		)
		return level == config.AccessControlLevelReadwrite
	}

	switch attributeVariant {
	case "name":
		names := []string{"name", "given_name", "family_name", "middle_name", "nickname"}
		for _, name := range names {
			editable := isEditable("/" + name)
			if editable {
				return true
			}
		}
		return false
	default:
		return isEditable("/" + attributeVariant)
	}
}

func (h *AuthflowV2SettingsProfileEditHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

		hasPermissionToEdit := h.isAttributeEditable(variant)

		if !hasPermissionToEdit {
			h.Renderer.RenderHTML(w, r, TemplateSettingsProfileNoPermission, data)
			return nil
		}

		h.Renderer.RenderHTML(w, r, settingsTemplate, data)
		return nil
	})

	ctrl.PostAction("save", func() error {
		userID := *session.GetUserID(r.Context())
		PatchGenderForm(r.Form)
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
