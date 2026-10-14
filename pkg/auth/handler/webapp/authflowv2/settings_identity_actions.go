package authflowv2

import (
	"context"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/session"
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

		canEdit := !updateDisabled
		canVerify := !verified && verificationEnabled
		canRemove := !deleteDisabled && identifiableCount > 1

		out[info.ID] = canEdit || canVerify || canRemove
	}
	return out, nil
}

// hasAtMostOneListItem reports whether the identity list page of a login ID
// type holds no item or one item. Such a list is not worth navigating to.
func hasAtMostOneListItem(all []*identity.Info, loginIDType model.LoginIDKeyType) bool {
	var claim string
	switch loginIDType {
	case model.LoginIDKeyTypeEmail:
		claim = stdattrs.Email
	case model.LoginIDKeyTypePhone:
		claim = stdattrs.PhoneNumber
	}

	count := 0
	for _, info := range all {
		switch info.Type {
		case model.IdentityTypeLoginID:
			if info.LoginID.LoginIDType == loginIDType {
				count++
			}
		case model.IdentityTypeOAuth:
			if v, _ := info.OAuth.Claims[claim].(string); claim != "" && v != "" {
				count++
			}
		}
	}
	return count <= 1
}

// SettingsListURLMakers are template functions to link to an identity list
// page. When the list holds at most one item, they return skipTo instead.
type SettingsListURLMakers struct {
	MakeListEmailURL    func(skipTo string, pairs ...string) string
	MakeListPhoneURL    func(skipTo string, pairs ...string) string
	MakeListUsernameURL func(skipTo string, pairs ...string) string
}

func NewSettingsListURLMakers(makeURL func(path string, pairs ...string) string, all []*identity.Info) SettingsListURLMakers {
	maker := func(loginIDType model.LoginIDKeyType, path string) func(string, ...string) string {
		return func(skipTo string, pairs ...string) string {
			if hasAtMostOneListItem(all, loginIDType) {
				return skipTo
			}
			return makeURL(path, pairs...)
		}
	}
	return SettingsListURLMakers{
		MakeListEmailURL:    maker(model.LoginIDKeyTypeEmail, "/settings/identity/email"),
		MakeListPhoneURL:    maker(model.LoginIDKeyTypePhone, "/settings/identity/phone"),
		MakeListUsernameURL: maker(model.LoginIDKeyTypeUsername, "/settings/identity/username"),
	}
}

// SettingsListURLService makes URLs to identity list pages for places that do
// not have the user's identities at hand.
type SettingsListURLService struct {
	Database   *appdb.Handle
	Identities SettingsIdentityService
}

// URL is for use outside a transaction, e.g. to redirect after a POST.
func (s *SettingsListURLService) URL(ctx context.Context, loginIDType model.LoginIDKeyType, loginIDKey string, skipTo string) (string, error) {
	var all []*identity.Info
	err := s.Database.WithTx(ctx, func(ctx context.Context) (err error) {
		all, err = s.Identities.ListByUser(ctx, *session.GetUserID(ctx))
		return err
	})
	if err != nil {
		return "", err
	}

	makeURL := func(path string, pairs ...string) string {
		u := url.URL{Path: path}
		q := u.Query()
		for i := 0; i+1 < len(pairs); i += 2 {
			q.Set(pairs[i], pairs[i+1])
		}
		u.RawQuery = q.Encode()
		return u.String()
	}
	makers := NewSettingsListURLMakers(makeURL, all)

	switch loginIDType {
	case model.LoginIDKeyTypeEmail:
		return makers.MakeListEmailURL(skipTo, "q_login_id_key", loginIDKey), nil
	case model.LoginIDKeyTypePhone:
		return makers.MakeListPhoneURL(skipTo, "q_login_id_key", loginIDKey), nil
	default:
		return makers.MakeListUsernameURL(skipTo, "q_login_id_key", loginIDKey), nil
	}
}

// Makers is for use outside a transaction, e.g. in GetData of a GET handler.
func (s *SettingsListURLService) Makers(ctx context.Context, makeURL func(path string, pairs ...string) string) (SettingsListURLMakers, error) {
	var makers SettingsListURLMakers
	err := s.Database.WithTx(ctx, func(ctx context.Context) (err error) {
		makers, err = s.MakersInTx(ctx, makeURL)
		return err
	})
	return makers, err
}

// MakersInTx is for use in a transaction.
func (s *SettingsListURLService) MakersInTx(ctx context.Context, makeURL func(path string, pairs ...string) string) (SettingsListURLMakers, error) {
	all, err := s.Identities.ListByUser(ctx, *session.GetUserID(ctx))
	if err != nil {
		return SettingsListURLMakers{}, err
	}
	return NewSettingsListURLMakers(makeURL, all), nil
}
