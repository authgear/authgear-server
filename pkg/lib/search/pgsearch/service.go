package pgsearch

import (
	"fmt"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/search/reindex"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Service struct {
	AppID     config.AppID
	Store     *Store
	Reindexer *reindex.Reindexer
}

func (s *Service) QueryUser(
	searchKeyword string,
	sortOption user.SortOption,
	pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error) {
	if s.Store == nil {
		return nil, fmt.Errorf("search database credential is not provided")
	}
	var refs []apimodel.PageItemRef
	err := s.Store.Database.ReadOnly(func() error {
		var err error
		refs, err = s.Store.QueryUser(searchKeyword, sortOption, pageArgs)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return refs, nil
}

func (s *Service) ReindexUser(userID string, isDelete bool) (err error) {
	return s.Reindexer.ReindexUser(config.SearchImplementationPostgresql, userID, isDelete)
}
