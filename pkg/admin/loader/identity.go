package loader

import (
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func EncodeIdentityID(ref *identity.Ref) string {
	return strings.Join([]string{
		ref.UserID,
		string(ref.Type),
		ref.ID,
	}, "|")
}

func DecodeIdentityID(id string) (*identity.Ref, error) {
	parts := strings.Split(id, "|")
	if len(parts) != 3 {
		return nil, apierrors.NewInvalid("invalid ID")
	}
	return &identity.Ref{
		UserID: parts[0],
		Type:   authn.IdentityType(parts[1]),
		Meta:   model.Meta{ID: parts[2]},
	}, nil
}

type IdentityLoaderIdentityService interface {
	GetMany(refs []*identity.Ref) ([]*identity.Info, error)
}

type IdentityLoader struct {
	*graphqlutil.DataLoader `wire:"-"`

	Identities IdentityLoaderIdentityService
}

func NewIdentityLoader(identities IdentityLoaderIdentityService) *IdentityLoader {
	l := &IdentityLoader{
		Identities: identities,
	}
	l.DataLoader = graphqlutil.NewDataLoader(l.LoadFunc)
	return l
}

func (l *IdentityLoader) LoadFunc(keys []interface{}) ([]interface{}, error) {
	// Prepare refs.
	refs := make([]*identity.Ref, len(keys))
	for i, key := range keys {
		ref, err := DecodeIdentityID(key.(string))
		if err != nil {
			return nil, err
		}
		refs[i] = ref
	}

	// Get entities.
	entities, err := l.Identities.GetMany(refs)
	if err != nil {
		return nil, err
	}

	// Create map.
	entityMap := make(map[string]*identity.Info)
	for _, entity := range entities {
		entityMap[EncodeIdentityID(entity.ToRef())] = entity
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
