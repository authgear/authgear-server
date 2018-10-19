// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package builder

import (
	"errors"
	"fmt"

	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
)

func SortOrderBySQL(alias string, sort record.Sort) (string, error) {
	var expr string

	switch sort.Expression.Type {
	case record.KeyPath:
		expr = fullQuoteIdentifier(alias, sort.Expression.Value.(string))
	case record.Function:
		var err error
		expr, err = funcOrderBySQL(alias, sort.Expression.Value.(record.Func))
		if err != nil {
			return "", err
		}
	default:
		return "", errors.New("invalid Sort: specify either KeyPath or Func")
	}

	order, err := sortOrderOrderBySQL(sort.Order)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(expr + " " + order), nil
}

// due to sq not being able to pass args in OrderBy, we can't re-use funcToSQLOperand
func funcOrderBySQL(alias string, fun record.Func) (string, error) {
	switch f := fun.(type) {
	case record.DistanceFunc:
		sql := fmt.Sprintf(
			"ST_Distance_Sphere(%s, ST_MakePoint(%f, %f))",
			fullQuoteIdentifier(alias, f.Field),
			f.Location.Lng(),
			f.Location.Lat(),
		)
		return sql, nil
	default:
		return "", fmt.Errorf("got unrecgonized record.Func = %T", fun)
	}
}

func sortOrderOrderBySQL(order record.SortOrder) (string, error) {
	switch order {
	case record.Asc:
		return "ASC", nil
	case record.Desc:
		return "DESC", nil
	default:
		return "", fmt.Errorf("unknown sort order = %v", order)
	}
}
