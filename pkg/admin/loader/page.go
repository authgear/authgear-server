package loader

import (
	"reflect"

	"github.com/graphql-go/relay"

	"github.com/authgear/authgear-server/pkg/admin/utils"
	"github.com/authgear/authgear-server/pkg/lib/api/model"
)

const MaxPageSize uint64 = 20

type PageArgs struct {
	Before model.PageCursor
	After  model.PageCursor
	First  *uint64
	Last   *uint64
}

func NewPageArgs(args relay.ConnectionArguments) PageArgs {
	pageArgs := PageArgs{
		Before: model.PageCursor(args.Before),
		After:  model.PageCursor(args.After),
	}

	var first, last *uint64
	if args.First >= 0 {
		value := uint64(args.First)
		if value > MaxPageSize {
			value = MaxPageSize
		}
		first = &value
	}
	if args.Last >= 0 {
		value := uint64(args.Last)
		if value > MaxPageSize {
			value = MaxPageSize
		}
		last = &value
	}
	if first == nil && last == nil {
		value := MaxPageSize
		first = &value
	}

	pageArgs.First = first
	pageArgs.Last = last
	return pageArgs
}

type TotalCountFunc func() (uint64, error)

type PageResult struct {
	HasPreviousPage bool
	HasNextPage     bool
	TotalCount      *utils.Lazy
	Values          []model.PageItem
}

func NewPageResult(args PageArgs, values []model.PageItem, totalCount *utils.Lazy) *PageResult {
	hasPreviousPage := true
	hasNextPage := true

	var limit *uint64
	var hasPage *bool
	if args.First != nil {
		limit = args.First
		hasPage = &hasNextPage
	} else if args.Last != nil {
		limit = args.Last
		hasPage = &hasPreviousPage
	}

	if limit != nil && uint64(reflect.ValueOf(values).Len()) < *limit {
		*hasPage = false
	}

	return &PageResult{
		HasPreviousPage: hasPreviousPage,
		HasNextPage:     hasNextPage,
		TotalCount:      totalCount,
		Values:          values,
	}
}
