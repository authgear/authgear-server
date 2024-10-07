package openapi

import (
	"net/url"
	"strconv"
)

type ListPersonsQuery struct {
	Skip    *int
	Take    *int
	Order   *OrderEnum
	OrderBy *PersonsOrderByEnum
	Search  *string
}

func (r ListPersonsQuery) ToQuery() url.Values {
	q := url.Values{}
	if r.Skip != nil {
		q.Set("skip", strconv.Itoa(*r.Skip))
	}
	if r.Take != nil {
		q.Set("take", strconv.Itoa(*r.Take))
	}
	if r.Order != nil {
		q.Set("order", string(*r.Order))
	}
	if r.OrderBy != nil {
		q.Set("orderBy", string(*r.OrderBy))
	}
	if r.Search != nil {
		q.Set("search", *r.Search)
	}
	return q
}
