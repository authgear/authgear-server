package stdattrs

import (
	"fmt"
	"sort"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type ClaimStore interface {
	ListByClaimName(userID string, claimName string) ([]*verification.Claim, error)
}

type ServiceNoEvent struct {
	UserProfileConfig *config.UserProfileConfig
	Identities        IdentityService
	UserQueries       UserQueries
	UserStore         UserStore
	ClaimStore        ClaimStore
	Transformer       Transformer
}

func (s *ServiceNoEvent) PopulateIdentityAwareStandardAttributes(userID string) (err error) {
	// Get all the identities this user has.
	identities, err := s.Identities.ListByUser(userID)
	if err != nil {
		return
	}

	// Sort the identities with newer ones ordered first.
	sort.SliceStable(identities, func(i, j int) bool {
		a := identities[i]
		b := identities[j]
		return a.CreatedAt.After(b.CreatedAt)
	})

	// Generate a list of emails, phone numbers and usernames belong to the user.
	var emails []string
	var phoneNumbers []string
	var preferredUsernames []string
	for _, iden := range identities {
		if email, ok := iden.Claims[stdattrs.Email].(string); ok && email != "" {
			emails = append(emails, email)
		}
		if phoneNumber, ok := iden.Claims[stdattrs.PhoneNumber].(string); ok && phoneNumber != "" {
			phoneNumbers = append(phoneNumbers, phoneNumber)
		}
		if preferredUsername, ok := iden.Claims[stdattrs.PreferredUsername].(string); ok && preferredUsername != "" {
			preferredUsernames = append(preferredUsernames, preferredUsername)
		}
	}

	user, err := s.UserQueries.GetRaw(userID)
	if err != nil {
		return
	}

	updated := false

	// Clear dangling standard attributes.
	clear := func(key string, allowedValues []string) {
		if value, ok := user.StandardAttributes[key].(string); ok {
			if !slice.ContainsString(allowedValues, value) {
				delete(user.StandardAttributes, key)
				updated = true
			}
		}
	}
	clear(stdattrs.Email, emails)
	clear(stdattrs.PhoneNumber, phoneNumbers)
	clear(stdattrs.PreferredUsername, preferredUsernames)

	// Populate standard attributes.
	populate := func(key string, allowedValues []string) {
		if _, ok := user.StandardAttributes[key].(string); !ok {
			if len(allowedValues) > 0 {
				user.StandardAttributes[key] = allowedValues[0]
				updated = true
			}
		}
	}
	populate(stdattrs.Email, emails)
	populate(stdattrs.PhoneNumber, phoneNumbers)
	populate(stdattrs.PreferredUsername, preferredUsernames)

	if updated {
		err = s.UserStore.UpdateStandardAttributes(userID, user.StandardAttributes)
		if err != nil {
			return
		}
	}

	return
}

func (s *ServiceNoEvent) UpdateStandardAttributes(role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error {
	// Remove derived attributes to avoid failing the validation.
	stdAttrs = stdattrs.T(stdAttrs).WithDerivedAttributesRemoved()

	err := stdattrs.Validate(stdattrs.T(stdAttrs))
	if err != nil {
		return err
	}

	rawUser, err := s.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	accessControl := s.UserProfileConfig.StandardAttributes.GetAccessControl()
	err = stdattrs.T(rawUser.StandardAttributes).CheckWrite(
		accessControl,
		role,
		stdattrs.T(stdAttrs),
	)
	if err != nil {
		return err
	}

	identities, err := s.Identities.ListByUser(userID)
	if err != nil {
		return err
	}

	ownedEmails := make(map[string]struct{})
	ownedPhoneNumbers := make(map[string]struct{})
	ownedPreferredUsernames := make(map[string]struct{})
	for _, iden := range identities {
		if email, ok := iden.Claims[stdattrs.Email].(string); ok && email != "" {
			ownedEmails[email] = struct{}{}
		}
		if phoneNumber, ok := iden.Claims[stdattrs.PhoneNumber].(string); ok && phoneNumber != "" {
			ownedPhoneNumbers[phoneNumber] = struct{}{}
		}
		if preferredUsername, ok := iden.Claims[stdattrs.PreferredUsername].(string); ok && preferredUsername != "" {
			ownedPreferredUsernames[preferredUsername] = struct{}{}
		}
	}

	check := func(key string, allowedValues map[string]struct{}) error {
		if value, ok := stdAttrs[key].(string); ok {
			_, allowed := allowedValues[value]
			if !allowed {
				return fmt.Errorf("unowned %v: %v", key, value)
			}
		}
		return nil
	}

	err = check(stdattrs.Email, ownedEmails)
	if err != nil {
		return err
	}

	err = check(stdattrs.PhoneNumber, ownedPhoneNumbers)
	if err != nil {
		return err
	}

	err = check(stdattrs.PreferredUsername, ownedPreferredUsernames)
	if err != nil {
		return err
	}

	err = s.UserStore.UpdateStandardAttributes(userID, stdAttrs)
	if err != nil {
		return err
	}

	// In case email/phone_number/preferred_username was removed, we add them back.
	err = s.PopulateIdentityAwareStandardAttributes(userID)
	if err != nil {
		return err
	}

	return nil
}

// DeriveStandardAttributes populates email_verified and phone_number_verified,
// if email or phone_number are found in attrs.
func (s *ServiceNoEvent) DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	for key, value := range attrs {
		value, err := s.Transformer.StorageFormToRepresentationForm(key, value)
		if err != nil {
			return nil, err
		}

		// Copy
		out[key] = value

		// Email
		if key == stdattrs.Email {
			verified := false
			if str, ok := value.(string); ok {
				claims, err := s.ClaimStore.ListByClaimName(userID, stdattrs.Email)
				if err != nil {
					return nil, err
				}
				for _, claim := range claims {
					if claim.Value == str {
						verified = true
					}
				}
			}
			out[stdattrs.EmailVerified] = verified
		}

		// Phone number
		if key == stdattrs.PhoneNumber {
			verified := false
			if str, ok := value.(string); ok {
				claims, err := s.ClaimStore.ListByClaimName(userID, stdattrs.PhoneNumber)
				if err != nil {
					return nil, err
				}
				for _, claim := range claims {
					if claim.Value == str {
						verified = true
					}
				}
			}
			out[stdattrs.PhoneNumberVerified] = verified
		}
	}

	// updated_at
	out[stdattrs.UpdatedAt] = updatedAt.Unix()

	accessControl := s.UserProfileConfig.StandardAttributes.GetAccessControl()
	out = stdattrs.T(out).ReadWithAccessControl(
		accessControl,
		role,
	).ToClaims()

	return out, nil
}
