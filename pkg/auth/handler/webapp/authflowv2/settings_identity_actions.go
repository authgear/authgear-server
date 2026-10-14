package authflowv2

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// LoginIDViewActions tells, per login ID identity ID, whether its view page
// (settings_identity_view_*.html) offers Edit, Verify or Remove. An identity
// without any of them should not be linked to that page.
//
// Keep in sync with the conditions in the view handlers and templates.
func LoginIDViewActions(
	ctx context.Context,
	all []*identity.Info,
	verifications SettingsVerificationService,
	loginIDConfig *config.LoginIDConfig,
	verificationConfig *config.VerificationConfig,
) (map[string]bool, error) {
	var loginIDs []*identity.Info
	for _, info := range all {
		if info.Type == model.IdentityTypeLoginID {
			loginIDs = append(loginIDs, info)
		}
	}

	statuses, err := verifications.GetVerificationStatuses(ctx, loginIDs)
	if err != nil {
		return nil, err
	}

	identifiableCount := len(identity.ApplyFilters(all, identity.KeepIdentifiable))

	out := map[string]bool{}
	for _, info := range loginIDs {
		updateDisabled := true
		deleteDisabled := true
		if keyConfig, ok := loginIDConfig.GetKeyConfig(info.LoginID.LoginIDKey); ok {
			updateDisabled = *keyConfig.UpdateDisabled
			deleteDisabled = *keyConfig.DeleteDisabled
		}

		claims := statuses[info.ID]
		verified := len(claims) > 0 && claims[0].Verified

		verificationEnabled := false
		switch info.LoginID.LoginIDType {
		case model.LoginIDKeyTypeEmail:
			verificationEnabled = *verificationConfig.Claims.Email.Enabled
		case model.LoginIDKeyTypePhone:
			verificationEnabled = *verificationConfig.Claims.PhoneNumber.Enabled
		}

		// The email view page hides Edit for a verified email.
		canEdit := !updateDisabled &&
			!(info.LoginID.LoginIDType == model.LoginIDKeyTypeEmail && verified)
		canVerify := !verified && verificationEnabled
		canRemove := !deleteDisabled && identifiableCount > 1

		out[info.ID] = canEdit || canVerify || canRemove
	}
	return out, nil
}
