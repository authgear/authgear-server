package search

import (
	"context"
	"fmt"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/search/pgsearch"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Service struct {
	SearchConfig               *config.SearchConfig
	ElasticsearchService       *elasticsearch.Service
	PGSearchService            *pgsearch.Service
	GlobalSearchImplementation config.GlobalSearchImplementation
}

func (s *Service) QueryUser(
	ctx context.Context,
	searchKeyword string,
	filterOptions user.FilterOptions,
	sortOption user.SortOption,
	pageArgs graphqlutil.PageArgs) ([]apimodel.PageItemRef, *Stats, error) {
	switch s.SearchConfig.GetImplementation(s.GlobalSearchImplementation) {
	case config.SearchImplementationElasticsearch:
		result, stats, err := s.ElasticsearchService.QueryUser(searchKeyword, filterOptions, sortOption, pageArgs)
		if err != nil {
			return nil, nil, err
		}
		return result, StatsFromElasticsearch(stats), nil
	case config.SearchImplementationPostgresql:
		result, err := s.PGSearchService.QueryUser(ctx, searchKeyword, filterOptions, sortOption, pageArgs)
		if err != nil {
			return nil, nil, err
		}
		return result, &Stats{}, nil
	case config.SearchImplementationNone:
		return nil, nil, ErrSearchDisabled
	}
	panic(fmt.Errorf("unknown search implementation: %s", s.SearchConfig.GetImplementation(s.GlobalSearchImplementation)))
}
