package handler

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	"github.com/skygeario/skygear-server/pkg/record/dependency/recordconv"
)

// QueryParser is a context for parsing raw query to record.Query
type QueryParser struct {
	UserID string
}

// sortFromRaw parses the specified structure into a Sort struct.
//
// The structure takes the following form:
//
//     [ _expression_ , _sort_order ]
//
// Expression supports key path type or the function type. Literal type is
// not supported.
//
// Sort Order only supports `"asc"` or `"desc"`.
func (parser *QueryParser) sortFromRaw(rawSort []interface{}, sort *record.Sort) {
	// Parse expression.
	expr := parser.parseExpression(rawSort[0])
	if expr.Type == record.Literal {
		panic(errors.New("sort does not support literal"))
	}

	// Parse sort order.
	var sortOrder record.SortOrder
	orderStr, _ := rawSort[1].(string)
	if orderStr == "" {
		panic(errors.New("empty sort order in sort descriptor"))
	}
	switch orderStr {
	case "asc":
		sortOrder = record.Asc
	case "desc":
		sortOrder = record.Desc
	default:
		panic(fmt.Errorf("unknown sort order: %v", orderStr))
	}

	sort.Expression = expr
	sort.Order = sortOrder
}

func (parser *QueryParser) sortsFromRaw(rawSorts []interface{}) []record.Sort {
	length := len(rawSorts)
	sorts := make([]record.Sort, length, length)

	for i := range rawSorts {
		sortSlice, _ := rawSorts[i].([]interface{})
		if len(sortSlice) != 2 {
			panic(fmt.Errorf("got len(sort descriptor) = %v, want 2", len(sortSlice)))
		}
		parser.sortFromRaw(sortSlice, &sorts[i])
	}

	return sorts
}

func (parser *QueryParser) predicateOperatorFromString(operatorString string) record.Operator {
	switch operatorString {
	case "and":
		return record.And
	case "or":
		return record.Or
	case "not":
		return record.Not
	case "eq":
		return record.Equal
	case "gt":
		return record.GreaterThan
	case "lt":
		return record.LessThan
	case "gte":
		return record.GreaterThanOrEqual
	case "lte":
		return record.LessThanOrEqual
	case "neq":
		return record.NotEqual
	case "like":
		return record.Like
	case "ilike":
		return record.ILike
	case "in":
		return record.In
	case "func":
		return record.Functional
	default:
		panic(fmt.Errorf("unrecognized operator = %s", operatorString))
	}
}

func (parser *QueryParser) predicateFromRaw(rawPredicate []interface{}) record.Predicate {
	if len(rawPredicate) < 2 {
		panic(fmt.Errorf("got len(predicate) = %v, want at least 2", len(rawPredicate)))
	}

	rawOperator, ok := rawPredicate[0].(string)
	if !ok {
		panic(fmt.Errorf("got predicate[0]'s type = %T, want string", rawPredicate[0]))
	}

	predicate := record.Predicate{
		Operator: parser.predicateOperatorFromString(rawOperator),
		Children: make([]interface{}, 0),
	}
	if predicate.Operator == record.Functional {
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

// parseExpression parses the specific structure into an Expression struct.
//
// Accepts one of the following types:
//
// * { "$type": "keypath", "$val": "_key_path_name_" }    // key path
// * [ "_func_name_" , _expression_1_ , _expression_2_ ]  // function
// * 42                                                   // literal
func (parser *QueryParser) parseExpression(i interface{}) record.Expression {
	switch v := i.(type) {
	case map[string]interface{}:
		var keyPath string
		if err := recordconv.MapFrom(i, (*recordconv.MapKeyPath)(&keyPath)); err == nil {
			if keyPath == "_owner" {
				keyPath = "_owner_id"
			}
			return record.Expression{
				Type:  record.KeyPath,
				Value: keyPath,
			}
		}
	case []interface{}:
		if len(v) > 0 {
			if f, err := parser.parseFunc(v); err == nil {
				return record.Expression{
					Type:  record.Function,
					Value: f,
				}
			}
		}
	}

	return record.Expression{
		Type:  record.Literal,
		Value: recordconv.ParseLiteral(i),
	}
}

func (parser *QueryParser) parseFunc(s []interface{}) (f record.Func, err error) {
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

func (parser *QueryParser) parseDistanceFunc(s []interface{}) (record.DistanceFunc, error) {
	emptyDistanceFunc := record.DistanceFunc{}
	if len(s) != 2 {
		return emptyDistanceFunc, fmt.Errorf("want 2 arguments for distance func, got %d", len(s))
	}

	var field string
	if err := recordconv.MapFrom(s[0], (*recordconv.MapKeyPath)(&field)); err != nil {
		return emptyDistanceFunc, fmt.Errorf("invalid key path: %v", err)
	}

	var location record.Location
	if err := recordconv.MapFrom(s[1], (*recordconv.MapLocation)(&location)); err != nil {
		return emptyDistanceFunc, fmt.Errorf("invalid location: %v", err)
	}

	return record.DistanceFunc{
		Field:    field,
		Location: location,
	}, nil
}

func (parser *QueryParser) parseUserRelationFunc(s []interface{}) (record.UserRelationFunc, error) {
	emptyUserRelationFunc := record.UserRelationFunc{}
	if len(s) != 2 {
		return emptyUserRelationFunc, fmt.Errorf("want 2 arguments for user relation func, got %d", len(s))
	}

	var field string
	if err := recordconv.MapFrom(s[0], (*recordconv.MapKeyPath)(&field)); err != nil {
		return emptyUserRelationFunc, fmt.Errorf("invalid key path: %v", err)
	}

	var relation recordconv.MapRelation
	if err := recordconv.MapFrom(s[1], (*recordconv.MapRelation)(&relation)); err != nil {
		return emptyUserRelationFunc, fmt.Errorf("invalid relation: %v", err)
	}

	return record.UserRelationFunc{
		KeyPath:           field,
		RelationName:      relation.Name,
		RelationDirection: relation.Direction,
		User:              parser.UserID,
	}, nil

}

func (parser *QueryParser) queryFromRaw(rawQuery map[string]interface{}, query *record.Query) (err skyerr.Error) {
	defer func() {
		// use panic to escape from inner error
		if r := recover(); r != nil {
			switch queryErr := r.(type) {
			case skyerr.Error:
				err = queryErr.(skyerr.Error)
				return
			case error:
				logrus.WithField("rawQuery", rawQuery).Debugln("failed to construct query")
				err = skyerr.NewErrorf(skyerr.InvalidArgument, "failed to construct query: %v", queryErr.Error())
			default:
				logrus.WithField("recovered", r).Errorln("panic recovered while constructing query")
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
		query.ComputedKeys = map[string]record.Expression{}
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
				fmt.Sprintf(`expecting "%s" to be an array`, key),
				[]string{key}))

		}
	}
}

func mapToQueryHookFunc(parser *QueryParser) mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() != reflect.Map {
			return data, nil
		}

		query := record.Query{}
		if to != reflect.TypeOf(query) {
			return data, nil
		}

		if err := parser.queryFromRaw(data.(map[string]interface{}), &query); err != nil {
			return nil, err
		}

		return query, nil
	}
}

// queryAccessVisitorPredicateStackEntry is an entry in the queryAccessVisitor
// predicate stack.
type queryAccessVisitorPredicateStackEntry struct {
	// Predicate is the predicate being check.
	Predicate record.Predicate

	// SimpleComparison is the result of the simple comparison check.
	// If the predicate is performing simple comparison, the access
	// is considered "discover" and "compare" instead of just "compare".
	SimpleComparison bool
}

// queryAccessVisitor checks a Query struct to determine if the client
// is authorized to perform this query.
type queryAccessVisitor struct {
	// FieldACL is the Field ACL settings.
	FieldACL record.FieldACL

	// Record Type is type of the record being queried.
	RecordType string

	// AuthInfo is the current logged in user.
	AuthInfo *authinfo.AuthInfo

	// ExpressionACLChecker is the helper struct for evaluating whether
	// the Field ACL settings allow access for an Expression.
	ExpressionACLChecker ExpressionACLChecker

	// pStack stores the stack of the Predicate being checked. Empty
	// if not checking any predicate.
	pStack []queryAccessVisitorPredicateStackEntry

	// err stores the error encountered (if any)
	err skyerr.Error

	// inSort stores true if the visitor is checking the expression in sorts.
	inSort bool
}

func (c *queryAccessVisitor) VisitQuery(p record.Query)    {}
func (c *queryAccessVisitor) EndVisitQuery(p record.Query) {}

func (c *queryAccessVisitor) VisitPredicate(p record.Predicate) {
	// For each predicate, determine if the predicate is for simple comparison.
	// A predicate performs simple comparison if the predicate operator
	// falls into a certain group. If the parent of a predicate is performing
	// non-simple comparison, the predicate in question is always performing
	// non-simple comparison.
	simple := func() bool {
		op := p.Operator
		simpleOp := op == record.Equal || op == record.In || op == record.And
		if len(c.pStack) > 0 {
			return c.pStack[len(c.pStack)-1].SimpleComparison && simpleOp
		}
		return simpleOp
	}()

	// Push the predicate and the result of SimpleComparison into stack.
	c.pStack = append(
		c.pStack,
		queryAccessVisitorPredicateStackEntry{
			Predicate:        p,
			SimpleComparison: simple,
		},
	)
}

func (c *queryAccessVisitor) EndVisitPredicate(p record.Predicate) {
	// Pop from stack.
	c.pStack = c.pStack[:len(c.pStack)-1]
}

func (c *queryAccessVisitor) VisitSort(sort record.Sort) {
	c.inSort = true
}

func (c *queryAccessVisitor) EndVisitSort(sort record.Sort) {
	c.inSort = false
}

func (c *queryAccessVisitor) VisitExpression(expr record.Expression) {
	if c.err != nil {
		return
	}

	var accessMode record.FieldAccessMode
	if len(c.pStack) > 0 {
		// The predicate stack is non-empty, that means we are checking
		// the predicate part of the query.
		if c.pStack[len(c.pStack)-1].SimpleComparison {
			accessMode = record.DiscoverOrCompareFieldAccessMode
		} else {
			accessMode = record.CompareFieldAccessMode
		}
	} else if c.inSort {
		// We are checking the Sorts part of the query.
		accessMode = record.CompareFieldAccessMode
	} else {
		// We are checking other parts of the query. For the time being
		// we are checking the ComputedKeys part of the query.
		accessMode = record.ReadFieldAccessMode
	}

	// When we have determined the access mode, check whether the expression
	// is allowed access.
	if err := c.ExpressionACLChecker.Check(expr, accessMode); err != nil {
		c.err = err
		return
	}
}

func (c *queryAccessVisitor) EndVisitExpression(expr record.Expression) {
	// do nothing
}

func (c *queryAccessVisitor) Error() skyerr.Error {
	return c.err
}

type ExpressionACLChecker struct {
	FieldACL    record.FieldACL
	RecordType  string
	AuthInfo    *authinfo.AuthInfo
	RecordStore record.Store
}

func (c *ExpressionACLChecker) Check(expr record.Expression, accessMode record.FieldAccessMode) skyerr.Error {
	switch expr.Type {
	case record.KeyPath:
		return c.checkKeyPath(expr.Value.(string), accessMode)
	case record.Function:
		return c.checkFunc(expr.Value.(record.Func), accessMode)
	case record.Literal:
		return nil
	default:
		panic("unsupported expression type")
	}
}

func (c *ExpressionACLChecker) checkFunc(fn record.Func, accessMode record.FieldAccessMode) skyerr.Error {
	if keyPathFn, ok := fn.(record.KeyPathFunc); ok {
		for _, keyPath := range keyPathFn.ReferencedKeyPaths() {
			if err := c.checkKeyPath(keyPath, accessMode); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *ExpressionACLChecker) checkKeyPath(keyPath string, accessMode record.FieldAccessMode) skyerr.Error {
	recordType := c.RecordType
	components := strings.Split(keyPath, ".")

	var fields []record.FieldType
	if len(components) > 1 {
		// Since the keypath is consists of multiple components, we have
		// to check the column types to find the Field ACL setting for all
		// referenced records.
		var err error
		fields, err = record.TraverseColumnTypes(c.RecordStore, recordType, keyPath)
		if err != nil {
			return skyerr.NewError(skyerr.RecordQueryInvalid, err.Error())
		}
	}

	for i, component := range components {
		if !strings.HasPrefix(component, "_") && !c.FieldACL.Accessible(
			recordType,
			component,
			accessMode,
			c.AuthInfo,
			nil,
		) {
			var msg string
			switch accessMode {
			case record.DiscoverOrCompareFieldAccessMode:
				msg = fmt.Sprintf(`Cannot query on field "%s" due to Field ACL, need to be discoverable or comparable`, component)
			case record.CompareFieldAccessMode:
				msg = fmt.Sprintf(`Cannot query on field "%s" due to Field ACL, need to be comparable`, component)
			case record.ReadFieldAccessMode:
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
