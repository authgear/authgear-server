package facade

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AuditLogQuery interface {
	Count(opts audit.QueryPageOptions) (uint64, error)
	QueryPage(opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error)
}

type AuditLogFacade struct {
	AuditLogQuery         AuditLogQuery
	Clock                 clock.Clock
	AuditLogFeatureConfig *config.AuditLogFeatureConfig
}

func (f *AuditLogFacade) QueryPage(opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	// bounded the from time, if retrieve days of audit log is configured in the feature config
	if *f.AuditLogFeatureConfig.RetrievalDays != -1 {
		days := *f.AuditLogFeatureConfig.RetrievalDays
		boundedByTime := f.Clock.NowUTC().Add(time.Duration(-days) * (24 * time.Hour))
		if opts.RangeFrom == nil || opts.RangeFrom.Before(boundedByTime) {
			opts.RangeFrom = &boundedByTime
		}
	}

	refs, err := f.AuditLogQuery.QueryPage(opts, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		return f.AuditLogQuery.Count(opts)
	})), nil
}
