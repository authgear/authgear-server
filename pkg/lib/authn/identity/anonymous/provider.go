package anonymous

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var KeyIDFormat = regexp.MustCompile(`^[-\w]{8,64}$`)

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

func (p *Provider) List(ctx context.Context, userID string) ([]*identity.Anonymous, error) {
	is, err := p.Store.List(ctx, userID)
	if err != nil {
		return nil, err
	}

	sortIdentities(is)
	return is, nil
}

func (p *Provider) Get(ctx context.Context, userID, id string) (*identity.Anonymous, error) {
	return p.Store.Get(ctx, userID, id)
}

func (p *Provider) GetByKeyID(ctx context.Context, keyID string) (*identity.Anonymous, error) {
	return p.Store.GetByKeyID(ctx, keyID)
}

func (p *Provider) GetMany(ctx context.Context, ids []string) ([]*identity.Anonymous, error) {
	return p.Store.GetMany(ctx, ids)
}

func (p *Provider) New(
	userID string,
	keyID string,
	key []byte,
) *identity.Anonymous {
	i := &identity.Anonymous{
		ID:     uuid.New(),
		UserID: userID,
		KeyID:  keyID,
		Key:    key,
	}
	return i
}

func (p *Provider) Create(ctx context.Context, i *identity.Anonymous) error {
	now := p.Clock.NowUTC()
	i.CreatedAt = now
	i.UpdatedAt = now
	return p.Store.Create(ctx, i)
}

func (p *Provider) Delete(ctx context.Context, i *identity.Anonymous) error {
	return p.Store.Delete(ctx, i)
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

		key, ok = set.Key(0)
		if !ok {
			err = fmt.Errorf("empty JWK set")
			return
		}
		keyID = key.KeyID()

		// The client does include alg in the JWK.
		// Fix it by copying alg in the header.
		if key.Algorithm().String() == "" {
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

func (p *Provider) ParseRequest(requestJWT string, i *identity.Anonymous) (*Request, error) {
	key, err := i.ToJWK()
	if err != nil {
		return nil, err
	}

	set := jwk.NewSet()
	_ = set.AddKey(key)

	payload, err := jws.Verify([]byte(requestJWT), jws.WithKeySet(set))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT: %w", err)
	}

	req := &Request{}
	err = json.Unmarshal(payload, req)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT payload: %w", err)
	}

	req.KeyID = i.KeyID
	req.Key = key
	return req, nil
}

func sortIdentities(is []*identity.Anonymous) {
	sort.Slice(is, func(i, j int) bool {
		return is[i].KeyID < is[j].KeyID
	})
}
