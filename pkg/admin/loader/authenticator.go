package loader

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AuthenticatorService interface {
	GetMany(ref []*authenticator.Ref) ([]*authenticator.Info, error)
	Count(userID string) (uint64, error)
	ListRefsByUsers(userIDs []string) ([]*authenticator.Ref, error)
}

type AuthenticatorLoader struct {
	Authenticators AuthenticatorService
	loader         *graphqlutil.DataLoader `wire:"-"`
	listLoader     *graphqlutil.DataLoader `wire:"-"`
}

func (l *AuthenticatorLoader) Get(ref *authenticator.Ref) *graphqlutil.Lazy {
	if l.loader == nil {
		l.loader = graphqlutil.NewDataLoader(func(keys []interface{}) ([]interface{}, error) {
			refs := make([]*authenticator.Ref, len(keys))
			for i, id := range keys {
				refs[i] = id.(*authenticator.Ref)
			}

			infos, err := l.Authenticators.GetMany(refs)
			if err != nil {
				return nil, err
			}

			infoMap := make(map[string]*authenticator.Info)
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

func (l *AuthenticatorLoader) List(userID string) *graphqlutil.Lazy {
	if l.listLoader == nil {
		l.listLoader = graphqlutil.NewDataLoader(func(keys []interface{}) ([]interface{}, error) {
			ids := make([]string, len(keys))
			for i, id := range keys {
				ids[i] = id.(string)
			}

			refs, err := l.Authenticators.ListRefsByUsers(ids)
			if err != nil {
				return nil, err
			}

			sort.Slice(refs, func(i, j int) bool {
				if refs[i].CreatedAt != refs[j].CreatedAt {
					return refs[i].CreatedAt.Before(refs[j].CreatedAt)
				}
				return refs[i].ID < refs[j].ID
			})

			refsMap := make(map[string][]*authenticator.Ref)
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
