package e2e

import (
	"os"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

func (c *End2End) GenerateIDToken(appID string, userID string) (string, error) {

	secretYaml, err := os.ReadFile("var/authgear.secrets.yaml")
	if err != nil {
		return "", err
	}

	secretConfig, err := config.ParsePartialSecret(secretYaml)
	if err != nil {
		return "", err
	}

	oauthKeySecrets := secretConfig.LookupData(config.OAuthKeyMaterialsKey).(*config.OAuthKeyMaterials)

	token := jwt.New()
	err = token.Set(jwt.SubjectKey, userID)
	if err != nil {
		return "", err
	}
	jwk, ok := oauthKeySecrets.Set.Key(0)
	if !ok {
		panic("Invalid jwk key in secrets")
	}
	signed, err := jwtutil.Sign(token, jwa.RS256, jwk)
	if err != nil {
		return "", err
	}
	return string(signed), err
}
