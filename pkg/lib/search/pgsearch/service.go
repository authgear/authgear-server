package pgsearch

import (
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/searchdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Service struct {
	SQLBuilder  *searchdb.SQLBuilder
	SQLExecutor *searchdb.SQLExecutor
}

func (s *Service) QueryUser(
	searchKeyword string,
	sortOption user.SortOption,
	pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, error) {
	return nil, nil
}
