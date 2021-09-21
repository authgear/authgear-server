package biometric

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
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

func (p *Provider) GetMany(ids []string) ([]*Identity, error) {
	return p.Store.GetMany(ids)
}

func (p *Provider) New(
	userID string,
	keyID string,
	key []byte,
	deviceInfo map[string]interface{},
) *Identity {
	i := &Identity{
		ID:         uuid.New(),
		UserID:     userID,
		KeyID:      keyID,
		Key:        key,
		DeviceInfo: deviceInfo,
	}
	return i
}

func (p *Provider) Create(i *Identity) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(i)
}

func (p *Provider) Delete(i *Identity) error {
	return p.Store.Delete(i)
}

func (p *Provider) ParseRequestUnverified(requestJWT string) (r *Request, err error) {
	compact := []byte(requestJWT)

	hdr, jwtToken, err := jwtutil.SplitWithoutVerify(compact)
	if err != nil {
		err = fmt.Errorf("invalid JWT: %w", err)
		return
	}

	err = jwt.Validate(jwtToken,
		jwt.WithClock(jwtClock{p.Clock}),
		jwt.WithAcceptableSkew(duration.ClockSkew),
	)
	if err != nil {
		err = fmt.Errorf("invalid JWT: %w", err)
		return
	}

	var key jwk.Key
	var keyID string
	if jwkIface, ok := hdr.Get("jwk"); ok {
		var jwkBytes []byte
		jwkBytes, err = json.Marshal(jwkIface)
		if err != nil {
			err = fmt.Errorf("invalid JWK: %w", err)
			return
		}

		var set jwk.Set
		set, err = jwk.Parse(jwkBytes)
		if err != nil {
			err = fmt.Errorf("invalid JWK: %w", err)
			return
		}

		key, ok = set.Get(0)
		if !ok {
			err = fmt.Errorf("empty JWK set")
			return
		}
		keyID = key.KeyID()

		// The client does include alg in the JWK.
		// Fix it by copying alg in the header.
		if key.Algorithm() == "" {
			_ = key.Set(jws.AlgorithmKey, hdr.Algorithm())
		}
	} else if kid := hdr.KeyID(); kid != "" {
		key = nil
		keyID = kid
	} else {
		err = errors.New("no key provided")
		return
	}

	typ := hdr.Type()
	if typ != RequestTokenType {
		err = errors.New("invalid JWT type")
		return
	}

	if !KeyIDFormat.MatchString(keyID) {
		err = errors.New("invalid key ID format")
		return
	}

	token, err := jws.ParseString(requestJWT)
	if err != nil {
		err = fmt.Errorf("invalid JWT: %w", err)
		return
	}

	var req Request
	err = json.Unmarshal(token.Payload(), &req)
	if err != nil {
		err = fmt.Errorf("invalid JWT payload: %w", err)
		return
	}

	req.Key = key
	req.KeyID = keyID
	r = &req
	return
}

func (p *Provider) ParseRequest(requestJWT string, identity *Identity) (*Request, error) {
	key, err := identity.toJWK()
	if err != nil {
		return nil, err
	}

	set := jwk.NewSet()
	_ = set.Add(key)

	payload, err := jws.VerifySet([]byte(requestJWT), set)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT: %w", err)
	}

	req := &Request{}
	err = json.Unmarshal(payload, req)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT payload: %w", err)
	}

	req.KeyID = identity.KeyID
	req.Key = key
	return req, nil
}

func sortIdentities(is []*Identity) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].CreatedAt.Before(is[j].CreatedAt)
	})
}
