package search

import (
	"fmt"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/search/pgsearch"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Service struct {
	SearchConfig         *config.SearchConfig
	ElasticsearchService *elasticsearch.Service
	PGSearchService      *pgsearch.Service
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
		return result, StatsFromElasticsearch(stats), nil
	case config.SearchImplementationPostgresql:
		// TODO(tung): Support filter options
		result, err := s.PGSearchService.QueryUser(searchKeyword, sortOption, pageArgs)
		if err != nil {
			return nil, nil, err
		}
		return result, &Stats{}, nil
	}
	return nil, nil, fmt.Errorf("unknown search implementation: %s", s.SearchConfig.GetImplementation())
}
