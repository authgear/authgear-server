package password

import (
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type LoginIDNormalizer interface {
	Normalize(loginID string) (string, error)
	ComputeUniqueKey(normalizeLoginID string) (string, error)
}

type LoginIDNormalizerFactory interface {
	NewNormalizer(loginIDKey string) LoginIDNormalizer
}

func NewLoginIDNormalizerFactory(
	loginIDsKeys []config.LoginIDKeyConfiguration,
	loginIDTypes *config.LoginIDTypesConfiguration,
) LoginIDNormalizerFactory {
	return &factoryImpl{
		loginIDsKeys: loginIDsKeys,
		loginIDTypes: loginIDTypes,
	}
}

type factoryImpl struct {
	loginIDsKeys []config.LoginIDKeyConfiguration
	loginIDTypes *config.LoginIDTypesConfiguration
}

func (f *factoryImpl) NewNormalizer(loginIDKey string) LoginIDNormalizer {
	for _, c := range f.loginIDsKeys {
		if c.Key == loginIDKey {
			return f.newNormalizer(c.Type)
		}
	}

	panic("password: invalid login id key: " + loginIDKey)
}

func (f *factoryImpl) newNormalizer(loginIDKeyType config.LoginIDKeyType) LoginIDNormalizer {
	metadataKey, _ := loginIDKeyType.MetadataKey()
	switch metadataKey {
	case metadata.Email:
		return &LoginIDEmailNormalizer{
			config: f.loginIDTypes.Email,
		}
	case metadata.Username:
		return &LoginIDUsernameNormalizer{
			config: f.loginIDTypes.Username,
		}
	}

	return &LoginIDNullNormalizer{}
}

type LoginIDEmailNormalizer struct {
	config *config.LoginIDTypeEmailConfiguration
}

func (n *LoginIDEmailNormalizer) Normalize(loginID string) (string, error) {
	parts := strings.Split(loginID, "@")
	local, domain := parts[0], parts[1]

	// convert the domain part
	c := cases.Fold()
	domain = c.String(domain)

	// convert the local part
	local = norm.NFKC.String(local)

	if !*n.config.CaseSensitive {
		local = c.String(local)
	}

	if *n.config.IgnoreDot {
		local = strings.Replace(local, ".", "", -1)
	}

	return local + "@" + domain, nil
}

func (n *LoginIDEmailNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	parts := strings.Split(normalizeLoginID, "@")
	local, domain := parts[0], parts[1]
	punycode, err := idna.ToASCII(domain)
	if err != nil {
		return "", err
	}
	domain = punycode
	return local + "@" + domain, nil
}

type LoginIDUsernameNormalizer struct {
	config *config.LoginIDTypeUsernameConfiguration
}

func (n *LoginIDUsernameNormalizer) Normalize(loginID string) (string, error) {
	loginID = norm.NFKC.String(loginID)

	c := cases.Fold()
	if !*n.config.CaseSensitive {
		loginID = c.String(loginID)
	}

	return loginID, nil
}

func (n *LoginIDUsernameNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	return normalizeLoginID, nil
}

type LoginIDNullNormalizer struct{}

func (n *LoginIDNullNormalizer) Normalize(loginID string) (string, error) {
	return loginID, nil
}

func (n *LoginIDNullNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	return normalizeLoginID, nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ LoginIDNormalizer = &LoginIDEmailNormalizer{}
	_ LoginIDNormalizer = &LoginIDUsernameNormalizer{}
	_ LoginIDNormalizer = &LoginIDNullNormalizer{}
)
