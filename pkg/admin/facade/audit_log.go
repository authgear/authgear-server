package facade

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AuditLogQuery interface {
	Count(ctx context.Context, opts audit.QueryPageOptions) (uint64, error)
	CountFraudProtectionDecisionRecords(ctx context.Context, opts audit.FraudProtectionDecisionRecordQueryOptions) (uint64, error)
	GetFraudProtectionDecisionRecordByID(ctx context.Context, id string) (*audit.FraudProtectionDecisionRecord, error)
	GetFraudProtectionOverview(ctx context.Context, opts audit.QueryPageOptions) (*audit.FraudProtectionOverview, error)
	QueryFraudProtectionDecisionRecordsPage(ctx context.Context, opts audit.FraudProtectionDecisionRecordQueryOptions, pageArgs graphqlutil.PageArgs) ([]*audit.FraudProtectionDecisionRecord, uint64, error)
	QueryPage(ctx context.Context, opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error)
}

type AuditLogFacade struct {
	AuditLogQuery         AuditLogQuery
	Clock                 clock.Clock
	AuditDatabase         *auditdb.ReadHandle
	AuditLogFeatureConfig *config.AuditLogFeatureConfig
}

func (f *AuditLogFacade) QueryPage(ctx context.Context, opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	f.boundRangeFrom(&opts)

	var refs []model.PageItemRef
	var count uint64
	var err error

	err = f.AuditDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		refs, err = f.AuditLogQuery.QueryPage(ctx, opts, pageArgs)
		if err != nil {
			return err
		}
		count, err = f.AuditLogQuery.Count(ctx, opts)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		return count, nil
	})), nil
}

func (f *AuditLogFacade) GetFraudProtectionOverview(ctx context.Context, queryOpts audit.QueryPageOptions) (*audit.FraudProtectionOverview, error) {
	queryOpts.ActivityTypes = []string{string(nonblocking.FraudProtectionDecisionRecorded)}
	f.boundRangeFrom(&queryOpts)

	var result *audit.FraudProtectionOverview
	err := f.AuditDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		var err error
		result, err = f.AuditLogQuery.GetFraudProtectionOverview(ctx, queryOpts)
		return err
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (f *AuditLogFacade) QueryFraudProtectionDecisionRecordsPage(
	ctx context.Context,
	opts audit.FraudProtectionDecisionRecordQueryOptions,
	pageArgs graphqlutil.PageArgs,
) ([]*audit.FraudProtectionDecisionRecord, uint64, *graphqlutil.PageResult, error) {
	f.boundFraudProtectionDecisionRecordRangeFrom(&opts)

	var items []*audit.FraudProtectionDecisionRecord
	var offset uint64
	var count uint64
	var err error

	err = f.AuditDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		items, offset, err = f.AuditLogQuery.QueryFraudProtectionDecisionRecordsPage(ctx, opts, pageArgs)
		if err != nil {
			return err
		}
		count, err = f.AuditLogQuery.CountFraudProtectionDecisionRecords(ctx, opts)
		return err
	})
	if err != nil {
		return nil, 0, nil, err
	}

	return items, offset, graphqlutil.NewPageResult(pageArgs, len(items), graphqlutil.NewLazy(func() (interface{}, error) {
		return count, nil
	})), nil
}

func (f *AuditLogFacade) GetFraudProtectionDecisionRecordByID(
	ctx context.Context,
	id string,
) (*audit.FraudProtectionDecisionRecord, error) {
	var item *audit.FraudProtectionDecisionRecord
	err := f.AuditDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		var err error
		item, err = f.AuditLogQuery.GetFraudProtectionDecisionRecordByID(ctx, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (f *AuditLogFacade) boundRangeFrom(opts *audit.QueryPageOptions) {
	// bounded the from time, if retrieve days of audit log is configured in the feature config
	if *f.AuditLogFeatureConfig.RetrievalDays != -1 {
		days := *f.AuditLogFeatureConfig.RetrievalDays
		boundedByTime := f.Clock.NowUTC().Add(time.Duration(-days) * (24 * time.Hour))
		if opts.RangeFrom == nil || opts.RangeFrom.Before(boundedByTime) {
			opts.RangeFrom = &boundedByTime
		}
	}
}

func (f *AuditLogFacade) boundFraudProtectionDecisionRecordRangeFrom(opts *audit.FraudProtectionDecisionRecordQueryOptions) {
	if *f.AuditLogFeatureConfig.RetrievalDays != -1 {
		days := *f.AuditLogFeatureConfig.RetrievalDays
		boundedByTime := f.Clock.NowUTC().Add(time.Duration(-days) * (24 * time.Hour))
		if opts.RangeFrom == nil || opts.RangeFrom.Before(boundedByTime) {
			opts.RangeFrom = &boundedByTime
		}
	}
}
