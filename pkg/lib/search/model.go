package search

import "github.com/authgear/authgear-server/pkg/lib/elasticsearch"

type Stats struct {
	TotalCount *int
}

func StatsFromElasticsearch(s *elasticsearch.Stats) *Stats {
	if s == nil {
		return nil
	}
	return &Stats{
		TotalCount: &s.TotalCount,
	}
}
