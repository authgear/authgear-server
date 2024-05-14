package stdattrs

import (
	"golang.org/x/text/language"

	"github.com/authgear/authgear-server/pkg/api/internalinterface"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/phone"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

//go:generate mockgen -source=normalizer.go -destination=normalizer_mock_test.go -package stdattrs

type LoginIDNormalizerFactory interface {
	NormalizerWithLoginIDType(loginIDKeyType model.LoginIDKeyType) internalinterface.LoginIDNormalizer
}

type Normalizer struct {
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func normalizeString(t T, key string) {
	if value, ok := t[key].(string); ok && value != "" {
		// noop
	} else {
		delete(t, key)
	}
}

func normalizeBool(t T, key string) {
	if _, ok := t[key].(bool); ok {
		// noop
	} else {
		delete(t, key)
	}
}

func (n *Normalizer) normalizeEmail(t T) error {
	if email, ok := t[Email].(string); ok && email != "" {
		emailNormalizer := n.LoginIDNormalizerFactory.NormalizerWithLoginIDType(model.LoginIDKeyTypeEmail)
		email, err := emailNormalizer.Normalize(email)
		if err != nil {
			return err
		}

		t[Email] = email
	} else {
		delete(t, Email)
	}

	return nil
}

func (n *Normalizer) normalizePhoneNumber(t T) error {
	if phoneNumber, ok := t[PhoneNumber].(string); ok && phoneNumber != "" {
		phoneNumber, err := phone.LegalParser.ParseInputPhoneNumber(phoneNumber)
		if err != nil {
			return err
		}

		t[PhoneNumber] = phoneNumber
	} else {
		delete(t, PhoneNumber)
	}

	return nil
}

func normalizeURL(t T, key string) {
	if value, ok := t[key].(string); ok && value != "" {
		err := validation.FormatURI{}.CheckFormat(value)
		if err != nil {
			delete(t, key)
		}
	} else {
		delete(t, key)
	}
}

func normalizeLocale(t T) {
	if locale, ok := t[Locale].(string); ok {
		tag, err := language.Parse(locale)
		if err != nil {
			delete(t, Locale)
		} else {
			// Use Canonical representation.
			t[Locale] = tag.String()
		}
	} else {
		delete(t, Locale)
	}
}

func normalizeZoneinfo(t T) {
	if value, ok := t[Zoneinfo].(string); ok {
		err := validation.FormatTimezone{}.CheckFormat(value)
		if err != nil {
			delete(t, Zoneinfo)
		}
	} else {
		delete(t, Zoneinfo)
	}
}

func normalizeBirthdate(t T) {
	if value, ok := t[Birthdate].(string); ok {
		err := validation.FormatBirthdate{}.CheckFormat(value)
		if err != nil {
			delete(t, Birthdate)
		}
	} else {
		delete(t, Birthdate)
	}
}

func normalizeAddress(t T) {
	if value, ok := t[Address].(map[string]interface{}); ok {
		normalizeString(T(value), Formatted)
		normalizeString(T(value), StreetAddress)
		normalizeString(T(value), Locality)
		normalizeString(T(value), Region)
		normalizeString(T(value), PostalCode)
		normalizeString(T(value), Country)
		if len(value) <= 0 {
			delete(t, Address)
		}
	} else {
		delete(t, Address)
	}
}

func (n *Normalizer) Normalize(t T) error {
	err := n.normalizeEmail(t)
	if err != nil {
		return err
	}

	err = n.normalizePhoneNumber(t)
	if err != nil {
		return err
	}

	normalizeString(t, Name)
	normalizeString(t, GivenName)
	normalizeString(t, FamilyName)
	normalizeString(t, MiddleName)
	normalizeString(t, Nickname)
	normalizeString(t, PreferredUsername)
	normalizeString(t, Gender)

	normalizeBool(t, EmailVerified)
	normalizeBool(t, PhoneNumberVerified)

	normalizeURL(t, Picture)
	normalizeURL(t, Profile)
	normalizeURL(t, Website)

	normalizeBirthdate(t)

	normalizeZoneinfo(t)

	normalizeLocale(t)

	normalizeAddress(t)

	return nil
}
