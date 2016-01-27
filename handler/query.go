package handler

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skydb/skyconv"
	"github.com/oursky/skygear/skyerr"
)

// QueryParser is a context for parsing raw query to skydb.Query
type QueryParser struct {
	UserID string
}

func (parser *QueryParser) sortFromRaw(rawSort []interface{}, sort *skydb.Sort) {
	var (
		keyPath   string
		funcExpr  skydb.Func
		sortOrder skydb.SortOrder
	)
	switch v := rawSort[0].(type) {
	case map[string]interface{}:
		if err := (*skyconv.MapKeyPath)(&keyPath).FromMap(v); err != nil {
			panic(err)
		}
	case []interface{}:
		var err error
		funcExpr, err = parser.parseFunc(v)
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("unexpected type of sort expression = %T", rawSort[0]))
	}

	orderStr, _ := rawSort[1].(string)
	if orderStr == "" {
		panic(errors.New("empty sort order in sort descriptor"))
	}
	switch orderStr {
	case "asc":
		sortOrder = skydb.Asc
	case "desc":
		sortOrder = skydb.Desc
	default:
		panic(fmt.Errorf("unknown sort order: %v", orderStr))
	}

	sort.KeyPath = keyPath
	sort.Func = funcExpr
	sort.Order = sortOrder
}

func (parser *QueryParser) sortsFromRaw(rawSorts []interface{}) []skydb.Sort {
	length := len(rawSorts)
	sorts := make([]skydb.Sort, length, length)

	for i := range rawSorts {
		sortSlice, _ := rawSorts[i].([]interface{})
		if len(sortSlice) != 2 {
			panic(fmt.Errorf("got len(sort descriptor) = %v, want 2", len(sortSlice)))
		}
		parser.sortFromRaw(sortSlice, &sorts[i])
	}

	return sorts
}

func (parser *QueryParser) predicateOperatorFromString(operatorString string) skydb.Operator {
	switch operatorString {
	case "and":
		return skydb.And
	case "or":
		return skydb.Or
	case "not":
		return skydb.Not
	case "eq":
		return skydb.Equal
	case "gt":
		return skydb.GreaterThan
	case "lt":
		return skydb.LessThan
	case "gte":
		return skydb.GreaterThanOrEqual
	case "lte":
		return skydb.LessThanOrEqual
	case "neq":
		return skydb.NotEqual
	case "like":
		return skydb.Like
	case "ilike":
		return skydb.ILike
	case "in":
		return skydb.In
	case "func":
		return skydb.Functional
	default:
		panic(fmt.Errorf("unrecognized operator = %s", operatorString))
	}
}

func (parser *QueryParser) predicateFromRaw(rawPredicate []interface{}) skydb.Predicate {
	if len(rawPredicate) < 2 {
		panic(fmt.Errorf("got len(predicate) = %v, want at least 2", len(rawPredicate)))
	}

	rawOperator, ok := rawPredicate[0].(string)
	if !ok {
		panic(fmt.Errorf("got predicate[0]'s type = %T, want string", rawPredicate[0]))
	}

	predicate := skydb.Predicate{
		Operator: parser.predicateOperatorFromString(rawOperator),
		Children: make([]interface{}, 0),
	}
	if predicate.Operator == skydb.Functional {
		predicate.Children = append(predicate.Children, parser.parseExpression(rawPredicate))
	} else if predicate.Operator.IsCompound() {
		for i := 1; i < len(rawPredicate); i++ {
			subRawPredicate, ok := rawPredicate[i].([]interface{})
			if !ok {
				panic(fmt.Errorf("got non-dict in subpredicate at %v", i-1))
			}
			predicate.Children = append(predicate.Children, parser.predicateFromRaw(subRawPredicate))
		}
	} else {
		for i := 1; i < len(rawPredicate); i++ {
			expr := parser.parseExpression(rawPredicate[i])
			if expr.Type == skydb.KeyPath && strings.Contains(expr.Value.(string), ".") {

				panic(fmt.Errorf("Key path `%s` is not supported.", expr.Value))
			}
			predicate.Children = append(predicate.Children, expr)
		}
	}

	if predicate.Operator.IsBinary() && len(predicate.Children) != 2 {
		panic(fmt.Errorf("Expected number of expressions be 2, got %v", len(predicate.Children)))
	}

	return predicate
}

func (parser *QueryParser) parseExpression(i interface{}) skydb.Expression {
	switch v := i.(type) {
	case map[string]interface{}:
		var keyPath string
		if err := skyconv.MapFrom(i, (*skyconv.MapKeyPath)(&keyPath)); err == nil {
			return skydb.Expression{
				Type:  skydb.KeyPath,
				Value: keyPath,
			}
		}
	case []interface{}:
		if len(v) > 0 {
			if f, err := parser.parseFunc(v); err == nil {
				return skydb.Expression{
					Type:  skydb.Function,
					Value: f,
				}
			}
		}
	}

	return skydb.Expression{
		Type:  skydb.Literal,
		Value: skyconv.ParseInterface(i),
	}
}

func (parser *QueryParser) parseFunc(s []interface{}) (f skydb.Func, err error) {
	keyword, _ := s[0].(string)
	if keyword != "func" {
		return nil, errors.New("not a function")
	}

	funcName, _ := s[1].(string)
	switch funcName {
	case "distance":
		f, err = parser.parseDistanceFunc(s[2:])
	case "userRelation":
		f, err = parser.parseUserRelationFunc(s[2:])
	case "":
		return nil, errors.New("empty function name")
	default:
		return nil, fmt.Errorf("got unrecgonized function name = %s", funcName)
	}

	return
}

func (parser *QueryParser) parseDistanceFunc(s []interface{}) (skydb.DistanceFunc, error) {
	emptyDistanceFunc := skydb.DistanceFunc{}
	if len(s) != 2 {
		return emptyDistanceFunc, fmt.Errorf("want 2 arguments for distance func, got %d", len(s))
	}

	var field string
	if err := skyconv.MapFrom(s[0], (*skyconv.MapKeyPath)(&field)); err != nil {
		return emptyDistanceFunc, fmt.Errorf("invalid key path: %v", err)
	}

	var location skydb.Location
	if err := skyconv.MapFrom(s[1], (*skyconv.MapLocation)(&location)); err != nil {
		return emptyDistanceFunc, fmt.Errorf("invalid location: %v", err)
	}

	return skydb.DistanceFunc{
		Field:    field,
		Location: location,
	}, nil
}

func (parser *QueryParser) parseUserRelationFunc(s []interface{}) (skydb.UserRelationFunc, error) {
	emptyUserRelationFunc := skydb.UserRelationFunc{}
	if len(s) != 2 {
		return emptyUserRelationFunc, fmt.Errorf("want 2 arguments for user relation func, got %d", len(s))
	}

	var field string
	if err := skyconv.MapFrom(s[0], (*skyconv.MapKeyPath)(&field)); err != nil {
		return emptyUserRelationFunc, fmt.Errorf("invalid key path: %v", err)
	}

	var relation skyconv.MapRelation
	if err := skyconv.MapFrom(s[1], (*skyconv.MapRelation)(&relation)); err != nil {
		return emptyUserRelationFunc, fmt.Errorf("invalid relation: %v", err)
	}

	return skydb.UserRelationFunc{
		KeyPath:           field,
		RelationName:      relation.Name,
		RelationDirection: relation.Direction,
		User:              parser.UserID,
	}, nil

}

func (parser *QueryParser) queryFromRaw(rawQuery map[string]interface{}, query *skydb.Query) (err skyerr.Error) {
	defer func() {
		// use panic to escape from inner error
		if r := recover(); r != nil {
			switch queryErr := r.(type) {
			case skyerr.Error:
				err = queryErr.(skyerr.Error)
				return
			case error:
				log.WithField("rawQuery", rawQuery).Debugln("failed to construct query")
				err = skyerr.NewErrorf(skyerr.InvalidArgument, "failed to construct query: %v", queryErr.Error())
			default:
				log.WithField("recovered", r).Errorln("panic recovered while constructing query")
				err = skyerr.NewError(skyerr.InvalidArgument, "error occurred while constructing query")
			}
		}
	}()
	recordType, _ := rawQuery["record_type"].(string)
	if recordType == "" {
		return skyerr.NewError(skyerr.InvalidArgument, "recordType cannot be empty")
	}
	query.Type = recordType

	mustDoSlice(rawQuery, "predicate", func(rawPredicate []interface{}) skyerr.Error {
		predicate := parser.predicateFromRaw(rawPredicate)
		if err := predicate.Validate(); err != nil {
			if err.Code() == skyerr.InternalQueryInvalid {
				return skyerr.NewInvalidArgument(
					fmt.Sprintf("query predicate is invalid: %v", err.Message()),
					[]string{"predicate"})
			}
			return err
		}
		query.Predicate = predicate
		return nil
	})

	mustDoSlice(rawQuery, "sort", func(rawSorts []interface{}) skyerr.Error {
		query.Sorts = parser.sortsFromRaw(rawSorts)
		return nil
	})

	if transientIncludes, ok := rawQuery["include"].(map[string]interface{}); ok {
		query.ComputedKeys = map[string]skydb.Expression{}
		for key, value := range transientIncludes {
			query.ComputedKeys[key] = parser.parseExpression(value)
		}
	}

	mustDoSlice(rawQuery, "desired_keys", func(desiredKeys []interface{}) skyerr.Error {
		query.DesiredKeys = make([]string, len(desiredKeys))
		for i, key := range desiredKeys {
			key, ok := key.(string)
			if !ok {
				return skyerr.NewError(skyerr.InvalidArgument, "unexpected value in desired_keys")
			}
			query.DesiredKeys[i] = key
		}
		return nil
	})

	if getCount, ok := rawQuery["count"].(bool); ok {
		query.GetCount = getCount
	}

	if offset, _ := rawQuery["offset"].(float64); offset > 0 {
		query.Offset = uint64(offset)
	}

	if limit, ok := rawQuery["limit"].(float64); ok {
		query.Limit = new(uint64)
		*query.Limit = uint64(limit)
	}
	return nil
}

// execute do when if the value of key in m is []interface{}. If value exists
// for key but its type is not []interface{} or do returns an error, it panics.
func mustDoSlice(m map[string]interface{}, key string, do func(value []interface{}) skyerr.Error) {
	vi, ok := m[key]
	if ok && vi != nil {
		v, ok := vi.([]interface{})
		if ok {
			if err := do(v); err != nil {
				panic(err)
			}
		} else {
			panic(skyerr.NewInvalidArgument(
				fmt.Sprintf(`expecting "%#s" to be an array`, key),
				[]string{key}))

		}
	}
}
