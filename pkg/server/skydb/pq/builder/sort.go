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

	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

func SortOrderBySQL(alias string, sort skydb.Sort) (string, error) {
	var expr string

	switch {
	case sort.KeyPath != "":
		expr = fullQuoteIdentifier(alias, sort.KeyPath)
	case sort.Func != nil:
		var err error
		expr, err = funcOrderBySQL(alias, sort.Func)
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
func funcOrderBySQL(alias string, fun skydb.Func) (string, error) {
	switch f := fun.(type) {
	case skydb.DistanceFunc:
		sql := fmt.Sprintf(
			"ST_Distance_Sphere(%s, ST_MakePoint(%f, %f))",
			fullQuoteIdentifier(alias, f.Field),
			f.Location.Lng(),
			f.Location.Lat(),
		)
		return sql, nil
	default:
		return "", fmt.Errorf("got unrecgonized skydb.Func = %T", fun)
	}
}

func sortOrderOrderBySQL(order skydb.SortOrder) (string, error) {
	switch order {
	case skydb.Asc:
		return "ASC", nil
	case skydb.Desc:
		return "DESC", nil
	default:
		return "", fmt.Errorf("unknown sort order = %v", order)
	}
}
