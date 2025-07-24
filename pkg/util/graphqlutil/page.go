package graphqlutil

import (
	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
)

const DEFAULT_MAX_PAGE_SIZE uint64 = 100

type Cursor string

type PageArgs struct {
	Before Cursor
	After  Cursor
	First  *uint64
	Last   *uint64
}

func NewPageArgsWithMaxPageSize(args relay.ConnectionArguments, maxPageSize uint64) PageArgs {
	pageArgs := PageArgs{
		Before: Cursor(args.Before),
		After:  Cursor(args.After),
	}

	var first, last *uint64
	if args.First >= 0 {
		value := uint64(args.First)
		if value > maxPageSize {
			value = maxPageSize
		}
		first = &value
	}
	if args.Last >= 0 {
		value := uint64(args.Last)
		if value > maxPageSize {
			value = maxPageSize
		}
		last = &value
	}
	if first == nil && last == nil {
		value := maxPageSize
		first = &value
	}

	pageArgs.First = first
	pageArgs.Last = last
	return pageArgs
}

func NewPageArgs(args relay.ConnectionArguments) PageArgs {
	return NewPageArgsWithMaxPageSize(args, DEFAULT_MAX_PAGE_SIZE)
}

type PageResult struct {
	HasPreviousPage bool
	HasNextPage     bool
	TotalCount      *Lazy
}

func NewPageResult(args PageArgs, itemsLen int, totalCount *Lazy) *PageResult {
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

	//nolint:gosec // G115
	itemsLen_uint64 := uint64(itemsLen)
	if limit != nil && itemsLen_uint64 < *limit {
		*hasPage = false
	}

	return &PageResult{
		HasPreviousPage: hasPreviousPage,
		HasNextPage:     hasNextPage,
		TotalCount:      totalCount,
	}
}
