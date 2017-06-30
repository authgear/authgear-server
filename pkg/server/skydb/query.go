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

//go:generate mockgen -package skydb -source=query.go -destination=mock_query_test.go

package skydb

import (
	"strings"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// SortOrder denotes an the order of Records returned from a Query.
type SortOrder int

// A list of SordOrder, their meaning is self descriptive.
const (
	Ascending SortOrder = iota
	Descending
	Asc  = Ascending
	Desc = Descending
)

// Sort specifies the order of a collection of Records returned from a Query.
//
// Record order can be sorted w.r.t. a record field or a value returned
// from a predefined function.
type Sort struct {
	Expression Expression
	Order      SortOrder
}

// Accept implements the Visitor pattern.
func (sort Sort) Accept(visitor Visitor) {
	if v, ok := visitor.(SortVisitor); ok {
		v.VisitSort(sort)
		defer v.EndVisitSort(sort)
	}
	if v, ok := visitor.(ExpressionVisitor); ok {
		sort.Expression.Accept(v)
	}
}

// Operator denotes how the result of a predicate is determined from
// its subpredicates or subexpressions.
//go:generate stringer -type=Operator
type Operator int

// A list of Operator.
const (
	And Operator = iota + 1
	Or
	Not
	Equal
	GreaterThan
	LessThan
	GreaterThanOrEqual
	LessThanOrEqual
	NotEqual
	Like
	ILike
	In
	Functional
)

// IsCompound checks whether the Operator is a compound operator, meaning the
// operator combine the results of other subpredicates.
func (op Operator) IsCompound() bool {
	switch op {
	default:
		return false
	case And, Or, Not:
		return true
	}
}

// IsBinary checks whether the Operator determines the result of a predicate
// by comparing two subexpressions.
func (op Operator) IsBinary() bool {
	switch op {
	default:
		return false
	case Equal, GreaterThan, LessThan, GreaterThanOrEqual, LessThanOrEqual, NotEqual, Like, ILike, In:
		return true
	}
}

// IsCommutative checks whether expressions on both side of the Operator
// can be swapped.
func (op Operator) IsCommutative() bool {
	switch op {
	default:
		return false
	case Equal, NotEqual:
		return true
	}
}

// ExpressionType is the type of an Expression.
type ExpressionType int

// A list of ExpressionTypes.
const (
	Literal ExpressionType = iota + 1
	KeyPath
	Function
)

// An Expression represents value to be compared against.
type Expression struct {
	Type  ExpressionType
	Value interface{}
}

func (expr Expression) IsEmpty() bool {
	return expr.Type == 0 && expr.Value == nil
}

func (expr Expression) IsKeyPath() bool {
	return expr.Type == KeyPath
}

func (expr Expression) IsLiteralString() bool {
	if expr.Type != Literal {
		return false
	}

	_, ok := expr.Value.(string)
	return ok
}

func (expr Expression) IsLiteralArray() bool {
	if expr.Type != Literal {
		return false
	}

	_, ok := expr.Value.([]interface{})
	return ok
}

func (expr Expression) IsLiteralMap() bool {
	if expr.Type != Literal {
		return false
	}

	_, ok := expr.Value.(map[string]interface{})
	return ok
}

func (expr Expression) IsLiteralNull() bool {
	if expr.Type != Literal {
		return false
	}

	return expr.Value == nil
}

func (expr Expression) KeyPathComponents() []string {
	if expr.Type != KeyPath {
		panic("expression is not a keypath")
	}

	return strings.Split(expr.Value.(string), ".")
}

// Accept implements the Visitor pattern.
func (expr Expression) Accept(visitor Visitor) {
	if v, ok := visitor.(ExpressionVisitor); ok {
		v.VisitExpression(expr)
		v.EndVisitExpression(expr)
	}
}

// Predicate is a representation of used in query for filtering records.
type Predicate struct {
	Operator Operator
	Children []interface{}
}

func (p Predicate) IsEmpty() bool {
	return p.Operator == 0 || p.Children == nil
}

// Validate returns an Error if a Predicate is invalid.
//
// If a Predicate is validated without error, nil is returned.
func (p Predicate) Validate() skyerr.Error {
	return p.validate(nil)
}

// validate is an internal version of the exported Validate() function.
//
// Additional information is passed as parameter to check the context
// in which the predicate is specified.
func (p Predicate) validate(parentPredicate *Predicate) skyerr.Error {
	if p.Operator.IsBinary() && len(p.Children) != 2 {
		return skyerr.NewErrorf(skyerr.RecordQueryInvalid,
			"binary predicate must have 2 operands, got %d", len(p.Children))
	}
	if p.Operator == Functional && len(p.Children) != 1 {
		return skyerr.NewErrorf(skyerr.RecordQueryInvalid,
			"functional predicate must have 1 operand, got %d", len(p.Children))
	}

	if p.Operator.IsCompound() {
		for _, child := range p.Children {
			predicate, ok := child.(Predicate)
			if !ok {
				return skyerr.NewError(skyerr.RecordQueryInvalid,
					"children of compound predicate must be a predicate")
			}

			if err := predicate.validate(&p); err != nil {
				return err
			}
		}
	} else {
		for _, child := range p.Children {
			_, ok := child.(Expression)
			if !ok {
				return skyerr.NewError(skyerr.RecordQueryInvalid,
					"children of simple predicate must be an expression")
			}
		}
	}

	switch p.Operator {
	case Functional:
		return p.validateFunctionalPredicate(parentPredicate)
	case Equal:
		return p.validateEqualPredicate(parentPredicate)
	}
	return nil
}

func (p Predicate) validateFunctionalPredicate(parentPredicate *Predicate) skyerr.Error {
	expr := p.Children[0].(Expression)
	if expr.Type != Function {
		return skyerr.NewError(skyerr.RecordQueryInvalid,
			`functional predicate must contain functional expression`)
	}

	switch f := expr.Value.(type) {
	case UserRelationFunc:
		if f.RelationName != "_friend" && f.RelationName != "_follow" {
			return skyerr.NewErrorf(skyerr.NotSupported,
				`user relation predicate with "%d" relation is not supported`,
				f.RelationName)
		}
	default:
		return skyerr.NewError(skyerr.NotSupported,
			`unsupported function for functional predicate`)
	}
	return nil
}

func (p Predicate) validateEqualPredicate(parentPredicate *Predicate) skyerr.Error {
	lhs := p.Children[0].(Expression)
	rhs := p.Children[1].(Expression)

	if lhs.IsLiteralMap() {
		return skyerr.NewErrorf(skyerr.NotSupported,
			`equal comparison of map "%v" is not supported`,
			lhs.Value)
	} else if lhs.IsLiteralArray() {
		return skyerr.NewErrorf(skyerr.NotSupported,
			`equal comparison of array "%v" is not supported`,
			lhs.Value)
	} else if rhs.IsLiteralMap() {
		return skyerr.NewErrorf(skyerr.NotSupported,
			`equal comparison of map "%v" is not supported`,
			rhs.Value)
	} else if rhs.IsLiteralArray() {
		return skyerr.NewErrorf(skyerr.NotSupported,
			`equal comparison of array "%v" is not supported`,
			rhs.Value)
	}
	return nil
}

// GetSubPredicates returns Predicate.Children as []Predicate.
//
// This method is only valid when Operator is either And, Or and Not. Caller
// is responsible to check for this preconditions. Otherwise the method
// will panic.
func (p Predicate) GetSubPredicates() (ps []Predicate) {
	for _, childPred := range p.Children {
		ps = append(ps, childPred.(Predicate))
	}
	return
}

// GetExpressions returns Predicate.Children as []Expression.
//
// This method is only valid when Operator is binary operator. Caller
// is responsible to check for this preconditions. Otherwise the method
// will panic.
func (p Predicate) GetExpressions() (ps []Expression) {
	for _, childPred := range p.Children {
		ps = append(ps, childPred.(Expression))
	}
	return
}

// Accept implements the Visitor pattern.
func (p Predicate) Accept(visitor Visitor) {
	if v, ok := visitor.(PredicateVisitor); ok {
		v.VisitPredicate(p)
		defer v.EndVisitPredicate(p)
	}
	if p.Operator.IsCompound() {
		for _, child := range p.Children {
			predicate, ok := child.(Predicate)
			if !ok {
				panic("children of compound predicate must be a predicate")
			}

			predicate.Accept(visitor)
		}
	} else {
		if v, ok := visitor.(ExpressionVisitor); ok {
			for _, child := range p.Children {
				expr, ok := child.(Expression)
				if !ok {
					panic("children of simple predicate must be an expression")
				}

				expr.Accept(v)
			}
		}
	}
}

// Query specifies the type, predicate and sorting order of Database
// query.
type Query struct {
	Type         string
	Predicate    Predicate
	Sorts        []Sort
	ComputedKeys map[string]Expression
	DesiredKeys  []string
	GetCount     bool
	Limit        *uint64
	Offset       uint64

	// The following fields are generated from the server side, rather
	// than supplied from the client side.
	ViewAsUser          *AuthInfo
	BypassAccessControl bool
}

// Accept implements the Visitor pattern.
func (q Query) Accept(visitor Visitor) {
	if v, ok := visitor.(QueryVisitor); ok {
		v.VisitQuery(q)
		defer v.EndVisitQuery(q)
	}

	if v, ok := visitor.(PredicateVisitor); ok {
		q.Predicate.Accept(v)
	}

	if v, ok := visitor.(SortVisitor); ok {
		for _, sort := range q.Sorts {
			sort.Accept(v)
		}
	}

	if v, ok := visitor.(ExpressionVisitor); ok {
		for _, expr := range q.ComputedKeys {
			expr.Accept(v)
		}
	}
}

// Func is a marker interface to denote a type being a function in skydb.
//
// skydb's function receives zero or more arguments and returns a DataType
// as a result. Result data type is currently omitted in this interface since
// skygear doesn't use it internally yet. In the future it can be utilized to
// provide more extensive type checking at handler level.
type Func interface {
	Args() []interface{}
	DataType() DataType
}

// KeyPathFunc is a marker interface to denote a func that
// references certain key paths.
type KeyPathFunc interface {
	// Returns a list of key paths that is referenced by this function.
	ReferencedKeyPaths() []string
}

// DistanceFunc represents a function that calculates distance between
// a user supplied location and a Record's field
type DistanceFunc struct {
	Field    string
	Location Location
}

// Args implements the Func interface
func (f DistanceFunc) Args() []interface{} {
	return []interface{}{f.Field, f.Location}
}

func (f DistanceFunc) DataType() DataType {
	return TypeNumber
}

// ReferencedKeyPaths implements the KeyPathFunc interface.
func (f DistanceFunc) ReferencedKeyPaths() []string {
	return []string{f.Field}
}

// CountFunc represents a function that count number of rows matching
// a query
type CountFunc struct {
	OverallRecords bool
}

// Args implements the Func interface
func (f CountFunc) Args() []interface{} {
	return []interface{}{}
}

func (f CountFunc) DataType() DataType {
	return TypeNumber
}

// UserRelationFunc represents a function that is used to evaulate
// whether a record satisfy certain user-based relation
type UserRelationFunc struct {
	KeyPath           string
	RelationName      string
	RelationDirection string
	User              string
}

// Args implements the Func interface
func (f UserRelationFunc) Args() []interface{} {
	return []interface{}{}
}

func (f UserRelationFunc) DataType() DataType {
	return TypeBoolean
}

// ReferencedKeyPaths implements the KeyPathFunc interface.
func (f UserRelationFunc) ReferencedKeyPaths() []string {
	return []string{f.KeyPath}
}

// Visitor is a marker interface
type Visitor interface{}

// FullQueryVisitor is a marker interface for all query-related visitors
type FullQueryVisitor interface {
	QueryVisitor
	PredicateVisitor
	SortVisitor
	ExpressionVisitor
}

// QueryVisitor is an interface that implements the Visitor pattern for
// the Query struct.
type QueryVisitor interface {
	VisitQuery(Query)
	EndVisitQuery(Query)
}

// PredicateVisitor is an interface that implements the Visitor pattern for
// the Predicate struct.
type PredicateVisitor interface {
	VisitPredicate(Predicate)
	EndVisitPredicate(Predicate)
}

// SortVisitor is an interface that implements the Visitor pattern for
// the Sort struct.
type SortVisitor interface {
	VisitSort(Sort)
	EndVisitSort(Sort)
}

// ExpressionVisitor is an interface that implements the Visitor pattern for
// the Expression struct.
type ExpressionVisitor interface {
	VisitExpression(Expression)
	EndVisitExpression(Expression)
}
