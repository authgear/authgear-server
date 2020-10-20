package facade

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func ConvertItems(items []model.PageItem) []graphqlutil.PageItem {
	out := make([]graphqlutil.PageItem, len(items))
	for i, item := range items {
		out[i] = graphqlutil.PageItem{
			Value:  item.Value,
			Cursor: graphqlutil.Cursor(item.Cursor),
		}
	}
	return out
}
