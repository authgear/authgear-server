package facade

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/audit"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type AuditLogQuery interface {
	Count(opts audit.QueryPageOptions) (uint64, error)
	QueryPage(opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, error)
}

type AuditLogFacade struct {
	AuditLogQuery AuditLogQuery
}

func (f *AuditLogFacade) QueryPage(opts audit.QueryPageOptions, pageArgs graphqlutil.PageArgs) ([]model.PageItemRef, *graphqlutil.PageResult, error) {
	refs, err := f.AuditLogQuery.QueryPage(opts, pageArgs)
	if err != nil {
		return nil, nil, err
	}

	return refs, graphqlutil.NewPageResult(pageArgs, len(refs), graphqlutil.NewLazy(func() (interface{}, error) {
		return f.AuditLogQuery.Count(opts)
	})), nil
}
