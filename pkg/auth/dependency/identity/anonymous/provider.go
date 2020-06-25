package anonymous

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
	"github.com/skygeario/skygear-server/pkg/jwtutil"
)

type jwtClock struct {
	Clock clock.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

type Provider struct {
	Store *Store
	Clock clock.Clock
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

func (p *Provider) ParseRequest(requestJWT string) (iden *Identity, r *Request, err error) {
	compact := []byte(requestJWT)
	var key jwk.Key

	hdr, jwtToken, err := jwtutil.SplitWithoutVerify(compact)
	if err != nil {
		err = fmt.Errorf("invalid JWT: %w", err)
		return
	}

	err = jwt.Verify(jwtToken,
		jwt.WithClock(jwtClock{p.Clock}),
		jwt.WithAcceptableSkew(5*time.Minute),
	)
	if err != nil {
		err = fmt.Errorf("invalid JWT: %w", err)
		return
	}

	if jwkIface, ok := hdr.Get("jwk"); ok {
		var jwkBytes []byte
		jwkBytes, err = json.Marshal(jwkIface)
		if err != nil {
			err = fmt.Errorf("invalid JWK: %w", err)
			return
		}

		var set *jwk.Set
		set, err = jwk.ParseBytes(jwkBytes)
		if err != nil {
			err = fmt.Errorf("invalid JWK: %w", err)
			return
		}

		key = set.Keys[0]

		iden, err = p.Store.GetByKeyID(key.KeyID())
		if err != nil && !errors.Is(err, identity.ErrIdentityNotFound) {
			return
		} else if err == nil {
			key, err = iden.toJWK()
			if err != nil {
				err = fmt.Errorf("invalid JWK: %w", err)
				return
			}
		}
	} else if kid := hdr.KeyID(); kid != "" {
		iden, err = p.Store.GetByKeyID(kid)
		if err != nil {
			err = fmt.Errorf("unknown key ID: %w", err)
			return
		}

		key, err = iden.toJWK()
		if err != nil {
			err = fmt.Errorf("invalid JWK: %w", err)
			return
		}
	} else {
		err = errors.New("no key provided")
		return
	}

	// The client does include alg in the JWK.
	// Fix it by copying alg in the header.
	if key.Algorithm() == "" {
		key.Set(jws.AlgorithmKey, hdr.Algorithm())
	}

	typ := hdr.Type()
	if typ != RequestTokenType {
		err = errors.New("invalid JWT type")
		return
	}

	if !KeyIDFormat.MatchString(key.KeyID()) {
		err = errors.New("invalid key ID format")
		return
	}

	payload, err := jws.VerifyWithJWK(compact, key)
	if err != nil {
		err = fmt.Errorf("invalid JWT signature: %w", err)
		return
	}

	var req Request
	err = json.Unmarshal(payload, &req)
	if err != nil {
		err = fmt.Errorf("invalid JWT payload: %w", err)
		return
	}

	req.Key = key
	r = &req
	return
}

func sortIdentities(is []*Identity) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].KeyID < is[j].KeyID
	})
}
