package stdattrs

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

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
	}

	return nil
}
