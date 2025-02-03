package identity

import (
	"time"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/util/phone"
)

type OAuth struct {
	ID                string                       `json:"id"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
	UserID            string                       `json:"user_id"`
	ProviderID        oauthrelyingparty.ProviderID `json:"provider_id"`
	ProviderSubjectID string                       `json:"provider_subject_id"`
	UserProfile       map[string]interface{}       `json:"user_profile,omitempty"`
	Claims            map[string]interface{}       `json:"claims,omitempty"`
	// This is a derived field and NOT persisted to database.
	// We still include it in JSON serialization so it can be persisted in the graph.
	ProviderAlias string `json:"provider_alias,omitempty"`
}

func (i *OAuth) ToInfo() *Info {
	return &Info{
		ID:        i.ID,
		UserID:    i.UserID,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
		Type:      model.IdentityTypeOAuth,

		OAuth: i,
	}
}

func (i *OAuth) IdentityAwareStandardClaims() map[model.ClaimName]string {
	claims := map[model.ClaimName]string{}
	if email, ok := i.Claims[string(model.ClaimEmail)].(string); ok {
		claims[model.ClaimEmail] = email
	}
	if phoneNumber, ok := i.Claims[string(model.ClaimPhoneNumber)].(string); ok {
		claims[model.ClaimPhoneNumber] = phoneNumber
	}
	if username, ok := i.Claims[string(model.ClaimPreferredUsername)].(string); ok {
		claims[model.ClaimPreferredUsername] = username
	}
	return claims
}

func (i *OAuth) GetDisplayName() string {
	if username, ok := i.Claims["preferred_username"].(string); ok && username != "" {
		// We don't know if username is a phone number or email, just try to mask it
		maskedMail := mail.MaskAddress(username)
		if maskedMail != "" {
			return maskedMail
		}
		maskedPhone := phone.Mask(username)
		if maskedPhone != "" {
			return maskedPhone
		}
		return username
	}

	if email, ok := i.Claims["email"].(string); ok && email != "" {
		return mail.MaskAddress(email)
	}

	if phoneNumber, ok := i.Claims["phone_number"].(string); ok && phoneNumber != "" {
		return phone.Mask(phoneNumber)
	}
	return ""
}

func (i *OAuth) Apple_MergeRawProfileAndClaims(rawProfile map[string]interface{}, claims map[string]interface{}) {
	// Use this heuristic to determine whether it is THE FIRST TIME authorization.
	_, isFirstTimeAuthorization := rawProfile[stdattrs.GivenName]
	if isFirstTimeAuthorization {
		i.UserProfile = rawProfile
		i.Claims = claims
	} else {
		// Otherwise we use the existing given_name and family_name.
		mergedRawProfile := make(map[string]interface{})
		mergedClaims := make(map[string]interface{})

		for k, v := range rawProfile {
			mergedRawProfile[k] = v
		}
		for k, v := range claims {
			mergedClaims[k] = v
		}

		merge := func(targetMap map[string]interface{}, sourceMap map[string]interface{}, key string) {
			if v, ok := sourceMap[key]; ok {
				targetMap[key] = v
			}
		}

		merge(mergedRawProfile, i.UserProfile, stdattrs.GivenName)
		merge(mergedRawProfile, i.UserProfile, stdattrs.FamilyName)

		merge(mergedClaims, i.Claims, stdattrs.GivenName)
		merge(mergedClaims, i.Claims, stdattrs.FamilyName)

		i.UserProfile = mergedRawProfile
		i.Claims = mergedClaims
	}
}
