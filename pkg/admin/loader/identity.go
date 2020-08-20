package loader

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type IdentityService interface {
	GetMany(ref []*identity.Ref) ([]*identity.Info, error)
	Count(userID string) (uint64, error)
	ListRefsByUsers(userIDs []string) ([]*identity.Ref, error)
}

type IdentityLoader struct {
	Identities IdentityService
	loader     *graphqlutil.DataLoader `wire:"-"`
	listLoader *graphqlutil.DataLoader `wire:"-"`
}

func (l *IdentityLoader) Get(ref *identity.Ref) *graphqlutil.Lazy {
	if l.loader == nil {
		l.loader = graphqlutil.NewDataLoader(func(keys []interface{}) ([]interface{}, error) {
			refs := make([]*identity.Ref, len(keys))
			for i, id := range keys {
				refs[i] = id.(*identity.Ref)
			}

			infos, err := l.Identities.GetMany(refs)
			if err != nil {
				return nil, err
			}

			infoMap := make(map[string]*identity.Info)
			for _, i := range infos {
				infoMap[i.ID] = i
			}
			values := make([]interface{}, len(keys))
			for i, ref := range refs {
				values[i] = infoMap[ref.ID]
			}
			return values, nil
		})
	}
	return l.loader.Load(ref)
}

func (l *IdentityLoader) List(userID string) *graphqlutil.Lazy {
	if l.listLoader == nil {
		l.listLoader = graphqlutil.NewDataLoader(func(keys []interface{}) ([]interface{}, error) {
			ids := make([]string, len(keys))
			for i, id := range keys {
				ids[i] = id.(string)
			}

			refs, err := l.Identities.ListRefsByUsers(ids)
			if err != nil {
				return nil, err
			}

			sort.Slice(refs, func(i, j int) bool {
				if refs[i].CreatedAt != refs[j].CreatedAt {
					return refs[i].CreatedAt.Before(refs[j].CreatedAt)
				}
				return refs[i].ID < refs[j].ID
			})

			refsMap := make(map[string][]*identity.Ref)
			for _, u := range refs {
				userRefs := refsMap[u.UserID]
				refsMap[u.UserID] = append(userRefs, u)
			}
			values := make([]interface{}, len(keys))
			for i, id := range ids {
				values[i] = refsMap[id]
			}
			return values, nil
		})
	}
	return l.listLoader.Load(userID)
}
