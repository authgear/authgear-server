package search

import (
	"fmt"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Service struct {
	SearchConfig         *config.SearchConfig
	ElasticsearchService *elasticsearch.Service
}

func (s *Service) QueryUser(
	searchKeyword string,
	filterOptions user.FilterOptions,
	sortOption user.SortOption,
	pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *Stats, error) {
	switch s.SearchConfig.GetImplementation() {
	case config.SearchImplementationElasticsearch:
		result, stats, err := s.ElasticsearchService.QueryUser(searchKeyword, filterOptions, sortOption, pageArgs)
		if err != nil {
			return nil, nil, err
		}
		return result, StatsFromElasticsearch(stats), err
	case config.SearchImplementationPostgresql:
		return nil, nil, fmt.Errorf("not implemented")
	}
	return nil, nil, fmt.Errorf("unknown search implementation: %s", s.SearchConfig.GetImplementation())
}
