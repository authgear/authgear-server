package stdattrs

import (
	"golang.org/x/text/language"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

//go:generate mockgen -source=normalizer.go -destination=normalizer_mock_test.go -package stdattrs

type LoginIDNormalizerFactory interface {
	NormalizerWithLoginIDType(loginIDKeyType config.LoginIDKeyType) loginid.Normalizer
}

type Normalizer struct {
	LoginIDNormalizerFactory LoginIDNormalizerFactory
}

func (n *Normalizer) Normalize(t T) error {
	if email, ok := t[Email].(string); ok && email != "" {
		emailNormalizer := n.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyTypeEmail)
		email, err := emailNormalizer.Normalize(email)
		if err != nil {
			return err
		}

		t[Email] = email
	} else {
		delete(t, Email)
	}

	// Verify locale is BCP47.
	if locale, ok := t[Locale].(string); ok {
		_, err := language.Parse(locale)
		if err != nil {
			delete(t, Locale)
		}
	} else {
		delete(t, Locale)
	}

	// Verify family_name, given_name, name are nonempty string
	if familyName, ok := t[FamilyName].(string); ok && familyName != "" {
		// noop
	} else {
		delete(t, FamilyName)
	}
	if givenName, ok := t[GivenName].(string); ok && givenName != "" {
		// noop
	} else {
		delete(t, GivenName)
	}
	if name, ok := t[Name].(string); ok && name != "" {
		// noop
	} else {
		delete(t, Name)
	}
	if nickname, ok := t[Nickname].(string); ok && nickname != "" {
		// noop
	} else {
		delete(t, Nickname)
	}

	// Verify picture and profile are URL.
	if picture, ok := t[Picture].(string); ok && picture != "" {
		err := validation.FormatURI{}.CheckFormat(picture)
		if err != nil {
			delete(t, Picture)
		}
	} else {
		delete(t, Picture)
	}
	if profile, ok := t[Profile].(string); ok && profile != "" {
		err := validation.FormatURI{}.CheckFormat(profile)
		if err != nil {
			delete(t, Profile)
		}
	} else {
		delete(t, Profile)
	}

	return nil
}
