package stdattrs

import (
	"fmt"
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type ServiceNoEvent struct {
	UserProfileConfig *config.UserProfileConfig
	Identities        IdentityService
	UserQueries       UserQueries
	UserStore         UserStore
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
