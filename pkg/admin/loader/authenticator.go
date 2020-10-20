package loader

import (
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func EncodeAuthenticatorID(ref *authenticator.Ref) string {
	return strings.Join([]string{
		ref.UserID,
		string(ref.Type),
		ref.ID,
	}, "|")
}

func DecodeAuthenticatorID(id string) (*authenticator.Ref, error) {
	parts := strings.Split(id, "|")
	if len(parts) != 3 {
		return nil, apierrors.NewInvalid("invalid ID")
	}
	return &authenticator.Ref{
		UserID: parts[0],
		Type:   authn.AuthenticatorType(parts[1]),
		Meta:   model.Meta{ID: parts[2]},
	}, nil
}

type AuthenticatorLoaderAuthenticatorService interface {
	GetMany(refs []*authenticator.Ref) ([]*authenticator.Info, error)
}

type AuthenticatorLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Authenticators AuthenticatorLoaderAuthenticatorService
}

func NewAuthenticatorLoader(authenticators AuthenticatorLoaderAuthenticatorService) *AuthenticatorLoader {
	l := &AuthenticatorLoader{
		Authenticators: authenticators,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *AuthenticatorLoader) LoadFunc(keys []interface{}) ([]interface{}, error) {
	// Prepare refs.
	refs := make([]*authenticator.Ref, len(keys))
	for i, key := range keys {
		ref, err := DecodeAuthenticatorID(key.(string))
		if err != nil {
			return nil, err
		}
		refs[i] = ref
	}

	// Get entities.
	entities, err := l.Authenticators.GetMany(refs)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*authenticator.Info)
	for _, entity := range entities {
		entityMap[EncodeAuthenticatorID(entity.ToRef())] = entity
	}

	// Ensure output is in correct order.
	out := make([]interface{}, len(keys))
	for i, key := range keys {
		entity := entityMap[key.(string)]
		if err != nil {
			out[i] = nil
		} else {
			out[i] = entity
		}
	}
	return out, nil
}
