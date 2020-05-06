package anonymous

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type Provider struct {
	Store *Store
}

func (p *Provider) List(userID string) ([]*Identity, error) {
	is, err := p.Store.List(userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) ListByClaim(name string, value string) ([]*Identity, error) {
	is, err := p.Store.ListByClaim(name, value)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) Get(userID, id string) (*Identity, error) {
	return p.Store.Get(userID, id)
}

func (p *Provider) GetByKeyID(keyID string) (*Identity, error) {
	return p.Store.GetByKeyID(keyID)
}

func (p *Provider) New(
	userID string,
	keyID string,
	key []byte,
) *Identity {
	i := &Identity{
		ID:     uuid.New(),
		UserID: userID,
		KeyID:  keyID,
		Key:    key,
	}
	return i
}

func (p *Provider) Create(i *Identity) error {
	return p.Store.Create(i)
}

func (p *Provider) Delete(i *Identity) error {
	return p.Store.Delete(i)
}

func (p *Provider) ParseRequest(requestJWT string) (*Identity, *Request, error) {
	var iden *Identity
	key := &jwk.RSAPublicKey{}
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Provided key material has higher priority than key ID
		if jwkMap, ok := token.Header["jwk"].(map[string]interface{}); ok {
			if err := key.ExtractMap(jwkMap); err != nil {
				return nil, fmt.Errorf("invalid JWK: %w", err)
			}

			var err error
			iden, err = p.Store.GetByKeyID(key.KeyID())
			if err != nil && !errors.Is(err, identity.ErrIdentityNotFound) {
				return nil, err
			}

			return key.Materialize()
		}
		if kid, ok := token.Header["kid"].(string); ok {
			var err error
			iden, err = p.Store.GetByKeyID(kid)
			if err != nil {
				return nil, fmt.Errorf("unknown key ID: %w", err)
			}

			var jwkMap map[string]interface{}
			if err := json.Unmarshal(iden.Key, &jwkMap); err != nil {
				return nil, fmt.Errorf("invalid JWK: %w", err)
			}
			if err := key.ExtractMap(jwkMap); err != nil {
				return nil, fmt.Errorf("invalid JWK: %w", err)
			}
			return key.Materialize()
		}

		return nil, errors.New("no key provided")
	}

	req := &Request{}
	token, err := jwt.ParseWithClaims(requestJWT, req, keyFunc)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid JWT signature: %w", err)
	}

	if typ, ok := token.Header["typ"].(string); !ok || typ != RequestTokenType {
		return nil, nil, errors.New("invalid JWT type")
	}

	req.Key = key
	return iden, req, nil
}

func sortIdentities(is []*Identity) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].KeyID < is[j].KeyID
	})
}
