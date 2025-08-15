package mockoidc

import (
	"encoding/json"
	"log"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const DefaultJWK = `
{
  "alg": "RS256",
  "d": "EJRXwvhxcS4P096nh5EGFofHGqn3UeZ1sq8RaXyBDDBGwbjMouTWM878BQOZSyHMzERo41HznmFixO3HO_Cdi3LxN7qBbwsrWQj3fbFxSNAea2CJyVdyZr5PAu47dzermMSFGXV1bZMeIH2w87LbVVXvAszxuyQgnExTMrsacSTmROhBJY9GY0D6R90aiglQtF9fnhU8tSG7WdSg1zh1bRTyW1pgs7qI2d1S4Km52FKZ38ABQPjYux3-zFvf3Az8JhwRLsXhv-G147uyaOcZiGcOykKMbZP5DTmsLdnzGW3K1De6Yg-copAbjsICTsjrIQzZaTUpMUaYjMcZZOtsAQ",
  "dp": "dnXy2iN-Brmp9IX-sh3saeEojfRqu3CeILZ6mXzJj9Lkvjf28gZWKWhwrr70r26jOY-MGbbsWEmtiZQlbwW81PztxJQL6gbn1gYqHCxYkZZ4EU0mpdQFM6Lpn7tNWVcNzEvQ0n8QccoT0bUHaZXj3jcznbpMXKQkTpTx0n9LMYE",
  "dq": "vr5iJF1Fvz-zyiicsjfaG97s9n78OVgiQoya2cIkEWut9UtX3LtT-UgwIkrkWE-uxBFcsxtWd47bPj2pd1LYd16Y2Orb_ocu-7x4HMtRAeIbdeFyNbxP7l_XzEmGrnRkxykCNhZ_Am1KvovSlnEVBRumrRJIKDBkKSm2uRBS6E8",
  "e": "AQAB",
  "ext": true,
  "key_ops": [
    "sign"
  ],
  "kty": "RSA",
  "n": "tbmGGgpT2Z3O5Y-X3bP_OjSlnE3Exb6yxrqkxrW2hlMnM5CKvfTheYr80IKiKdn6KbPFJLPqOtqT1As1GISefLLKRSShh9XV2C2lctpY8kpuX3M9UrzAjmHL2ld6yYlVlJcqdFxu1SPH2BarC72CcKzRCZ93Vn9fccqLki8CgtJBdW-zHNYvOAf-IyqY4z9fw5J8QK-ecSqEoCm8birjbHTcoQc83TuzHo1QYJE7gRJnV_bw4RV1VAAd3RVxDdOHTh576s_7px8Qj5TQm1tFjWpnzDhfu8PnOkUD9-GA4LffsHGWVoLD4mtPA0L3k4gKE30yn0YABr1ROwHMOHWJTw",
  "p": "3aBGRkqDiAd9Bum761Zie0ED7NwuXG1W5vAbeliu5mkfawJcl7HVswWOBKUvtahpLbIFW9NRtmiMGp0OiXF2mZGrRGbXkpNfYGDRUZ6Lr17xfA4l4JhSQa9yEXzQdW2g7riUXqcVNtBgyS5xf4FZFecUExpARgsz6-XApmWuekE",
  "q": "0ej0R9q7FwRihE6O1ytI8XzeIMx96y_xDVV-fZYUWR6QuVSvVdqy8KID41ltrX2z4uPKfe0RFSI-Fh_9VsvVn9n3Urog8Upnac0lVLqjxtc9xJH3fmEmJS7GWc4seWQeFsPUGQ5KDaDOY1qS8pdJLW1afq_C-B0_8yERwK0Zf48",
  "qi": "1007YwPI5BbduSirj-M2H9t68c90zO08IvhucmgR8fw1PPzKilzijdEwL5Bb9IJUDYYPrRh7ydQgx7NoK-mzHq4gtyeMnTN7KIq4X4dUIzURvBSkAZQr7Mg8xdVbmBJAHDXDnci13TCQgO3n39TKkXtV9VmkFvtSfSfTa2nEBXc",
  "kid": "FDKus3nofZY6iTkdOZLU86hKo1eIyHyeOtgMcsnleGw=",
  "use": "sig"
}
`

type Keypair struct {
	PrivateKey jwk.Key
	PublicKey  jwk.Key
}

func NewKeypair(key jwk.Key) (*Keypair, error) {
	if key == nil {
		return DefaultKeypair()
	}
	public, err := key.PublicKey()
	if err != nil {
		log.Fatalf("Failed to get public JWK: %v", err)
	}
	return &Keypair{
		PrivateKey: key,
		PublicKey:  public,
	}, nil
}

func DefaultKeypair() (*Keypair, error) {
	key, err := jwk.ParseKey([]byte(DefaultJWK))
	if err != nil {
		log.Fatalf("Failed to unmarshal JWK: %v", err)
	}

	public, err := key.PublicKey()
	if err != nil {
		log.Fatalf("Failed to get public JWK: %v", err)
	}
	return &Keypair{
		PrivateKey: key,
		PublicKey:  public,
	}, nil
}

func (k *Keypair) JWKKeySet() (jwk.Set, error) {
	key := k.PublicKey

	keySet := jwk.NewSet()
	err := keySet.AddKey(key)
	if err != nil {
		return nil, err
	}

	return keySet, nil
}

func (k *Keypair) JWKS() ([]byte, error) {
	keySet, err := k.JWKKeySet()
	if err != nil {
		return nil, err
	}

	return json.Marshal(keySet)
}

func (k *Keypair) SignJWT(token jwt.Token) (string, error) {
	key := k.PrivateKey

	compact, err := jwt.Sign(token, jwt.WithKey(key.Algorithm(), key))
	if err != nil {
		return "", err
	}

	return string(compact), nil
}

func (k *Keypair) VerifyJWT(token string, nowFunc func() time.Time) (jwt.Token, error) {
	keySet, err := k.JWKKeySet()
	if err != nil {
		return nil, err
	}

	return jwt.Parse([]byte(token),
		jwt.WithKeySet(keySet),
		jwt.WithClock(jwt.ClockFunc(nowFunc)),
	)
}
