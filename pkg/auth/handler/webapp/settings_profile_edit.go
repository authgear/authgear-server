package webapp

import (
	"context"

	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebSettingsProfileEditHTML = template.RegisterHTML(
	"web/settings_profile_edit.html",
	Components...,
)

type SettingsProfileEditUserService interface {
	GetRaw(ctx context.Context, id string) (*user.User, error)
}

type SettingsProfileEditStdAttrsService interface {
	UpdateStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
}

type SettingsProfileEditCustomAttrsService interface {
	UpdateCustomAttributesWithForm(ctx context.Context, role accesscontrol.Role, userID string, jsonPointerMap map[string]string) error
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
