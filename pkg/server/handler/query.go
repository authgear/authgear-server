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

package handler

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// QueryParser is a context for parsing raw query to skydb.Query
type QueryParser struct {
	UserID string
}

func (parser *QueryParser) sortFromRaw(rawSort []interface{}, sort *skydb.Sort) {
	var (
		expr      skydb.Expression
		sortOrder skydb.SortOrder
	)
	switch v := rawSort[0].(type) {
	case map[string]interface{}:
		var keyPath string
		if err := (*skyconv.MapKeyPath)(&keyPath).FromMap(v); err != nil {
			panic(err)
		}
		expr = skydb.Expression{
			Type:  skydb.KeyPath,
			Value: keyPath,
		}
	case []interface{}:
		var funcExpr skydb.Func
		var err error
		funcExpr, err = parser.parseFunc(v)
		if err != nil {
			panic(err)
		}
		expr = skydb.Expression{
			Type:  skydb.Function,
			Value: funcExpr,
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

	sort.Expression = expr
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
			if keyPath == "_owner" {
				keyPath = "_owner_id"
			}
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
		Value: skyconv.ParseLiteral(i),
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
	case "userDiscover":
		f, err = parser.parseUserDiscoverFunc(s[2:])
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

func (parser *QueryParser) parseUserDiscoverFunc(s []interface{}) (skydb.UserDiscoverFunc, error) {
	emptyUserDiscoverFunc := skydb.UserDiscoverFunc{}
	if len(s) != 1 {
		return emptyUserDiscoverFunc, fmt.Errorf("want 1 arguments for user discover func, got %d", len(s))
	}

	userData := struct {
		Usernames []string `mapstructure:"usernames"`
		Emails    []string `mapstructure:"emails"`
	}{}

	if err := mapstructure.Decode(s[0], &userData); err != nil {
		return emptyUserDiscoverFunc, err
	}

	return skydb.UserDiscoverFunc{
		Usernames: userData.Usernames,
		Emails:    userData.Emails,
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

func mapToQueryHookFunc(parser *QueryParser) mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() != reflect.Map {
			return data, nil
		}

		query := skydb.Query{}
		if to != reflect.TypeOf(query) {
			return data, nil
		}

		if err := parser.queryFromRaw(data.(map[string]interface{}), &query); err != nil {
			return nil, err
		}

		return query, nil
	}
}

type queryAccessVisitorPredicateStackEntry struct {
	Predicate        skydb.Predicate
	SimpleComparison bool
}

type queryAccessVisitor struct {
	pStack               []queryAccessVisitorPredicateStackEntry
	err                  skyerr.Error
	FieldACL             skydb.FieldACL
	RecordType           string
	UserInfo             *skydb.UserInfo
	ExpressionACLChecker ExpressionACLChecker
	inSort               bool
}

func (c *queryAccessVisitor) VisitQuery(p skydb.Query)    {}
func (c *queryAccessVisitor) EndVisitQuery(p skydb.Query) {}

func (c *queryAccessVisitor) VisitPredicate(p skydb.Predicate) {
	simple := func() bool {
		op := p.Operator
		simpleOp := op == skydb.Equal || op == skydb.In || op == skydb.And
		if len(c.pStack) > 0 {
			return c.pStack[len(c.pStack)-1].SimpleComparison && simpleOp
		}
		return simpleOp
	}()
	c.pStack = append(
		c.pStack,
		queryAccessVisitorPredicateStackEntry{
			Predicate:        p,
			SimpleComparison: simple,
		},
	)
}

func (c *queryAccessVisitor) EndVisitPredicate(p skydb.Predicate) {
	log.Infof("EndVisitPredicate: len: %d", len(c.pStack))
	c.pStack = c.pStack[:len(c.pStack)-1]
}

func (c *queryAccessVisitor) VisitSort(sort skydb.Sort) {
	c.inSort = true
}

func (c *queryAccessVisitor) EndVisitSort(sort skydb.Sort) {
	c.inSort = false
}

func (c *queryAccessVisitor) VisitExpression(expr skydb.Expression) {
	if c.err != nil {
		return
	}

	var accessMode skydb.FieldAccessMode
	if len(c.pStack) > 0 {
		if c.pStack[len(c.pStack)-1].SimpleComparison {
			accessMode = skydb.DiscoverOrCompareFieldAccessMode
		} else {
			accessMode = skydb.CompareFieldAccessMode
		}
	} else if c.inSort {
		accessMode = skydb.CompareFieldAccessMode
	} else {
		accessMode = skydb.ReadFieldAccessMode
	}

	if err := c.ExpressionACLChecker.Check(expr, accessMode); err != nil {
		c.err = err
		return
	}
}

func (c *queryAccessVisitor) EndVisitExpression(expr skydb.Expression) {
	// do nothing
}

func (c *queryAccessVisitor) Error() skyerr.Error {
	return c.err
}

type ExpressionACLChecker struct {
	FieldACL   skydb.FieldACL
	RecordType string
	UserInfo   *skydb.UserInfo
	Database   skydb.Database
}

func (c *ExpressionACLChecker) Check(expr skydb.Expression, accessMode skydb.FieldAccessMode) skyerr.Error {
	switch expr.Type {
	case skydb.KeyPath:
		return c.checkKeyPath(expr.Value.(string), accessMode)
	case skydb.Function:
		return c.checkFunc(expr.Value.(skydb.Func), accessMode)
	case skydb.Literal:
		return nil
	default:
		panic("unsupported expression type")
	}
}

func (c *ExpressionACLChecker) checkFunc(fn skydb.Func, accessMode skydb.FieldAccessMode) skyerr.Error {
	if keyPathFn, ok := fn.(skydb.KeyPathFunc); ok {
		for _, keyPath := range keyPathFn.ReferencedKeyPaths() {
			if err := c.checkKeyPath(keyPath, accessMode); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *ExpressionACLChecker) checkKeyPath(keyPath string, accessMode skydb.FieldAccessMode) skyerr.Error {
	recordType := c.RecordType
	components := strings.Split(keyPath, ".")

	var fields []skydb.FieldType
	if len(components) > 1 {
		// Since the keypath is consists of multiple components, we have
		// to check the column types to find the Field ACL setting for all
		// referenced records.
		var err error
		fields, err = skydb.TraverseColumnTypes(c.Database, recordType, keyPath)
		if err != nil {
			return skyerr.NewError(skyerr.RecordQueryInvalid, err.Error())
		}
	}

	for i, component := range components {
		if !strings.HasPrefix(component, "_") && !c.FieldACL.Accessible(
			recordType,
			component,
			accessMode,
			c.UserInfo,
			nil,
		) {
			var msg string
			switch accessMode {
			case skydb.DiscoverOrCompareFieldAccessMode:
				msg = fmt.Sprintf(`Cannot query on field "%s" due to Field ACL, need to be discoverable or comparable`, component)
			case skydb.CompareFieldAccessMode:
				msg = fmt.Sprintf(`Cannot query on field "%s" due to Field ACL, need to be comparable`, component)
			case skydb.ReadFieldAccessMode:
				msg = fmt.Sprintf(`Cannot query on field "%s" due to Field ACL, need to be readable`, component)
			}
			return skyerr.NewError(skyerr.RecordQueryDenied, msg)
		}

		isLast := (i == len(components)-1)
		if !isLast {
			if len(fields) <= i {
				panic("number of components in keypath does not match that in database schema")
			}
			recordType = fields[i].ReferenceType
		}
	}
	return nil
}
