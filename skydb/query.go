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

package skydb

import (
	"fmt"

	"github.com/skygeario/skygear-server/skyerr"
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
	KeyPath string
	Func    Func
	Order   SortOrder
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

// validates is an internal version of the exported Validate() function.
//
// Additional information is passed as parameter to check the context
// in which the predicate is specified.
func (p Predicate) validate(parentPredicate *Predicate) skyerr.Error {
	if p.Operator.IsBinary() && len(p.Children) != 2 {
		return skyerr.NewErrorf(skyerr.InternalQueryInvalid,
			"binary predicate must have 2 operands, got %d", len(p.Children))
	}
	if p.Operator == Functional && len(p.Children) != 1 {
		return skyerr.NewErrorf(skyerr.InternalQueryInvalid,
			"functional predicate must have 1 operand, got %d", len(p.Children))
	}

	if p.Operator.IsCompound() {
		for _, child := range p.Children {
			predicate, ok := child.(Predicate)
			if !ok {
				return skyerr.NewError(skyerr.InternalQueryInvalid,
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
				return skyerr.NewError(skyerr.InternalQueryInvalid,
					"children of simple predicate must be an expression")
			}
		}
	}

	switch p.Operator {
	case In:
		return p.validateInPredicate(parentPredicate)
	case Functional:
		return p.validateFunctionalPredicate(parentPredicate)
	case Equal:
		return p.validateEqualPredicate(parentPredicate)
	}
	return nil
}

func (p Predicate) validateInPredicate(parentPredicate *Predicate) skyerr.Error {
	lhs := p.Children[0].(Expression)
	rhs := p.Children[1].(Expression)

	if lhs.IsKeyPath() == rhs.IsKeyPath() {
		return skyerr.NewError(skyerr.InternalQueryInvalid,
			`either one of the operands of "IN" must be key path`)
	}

	if rhs.IsKeyPath() && !lhs.IsLiteralString() {
		return skyerr.NewError(skyerr.InternalQueryInvalid,
			`left operand of "IN" must be a string if comparing with a keypath`)
	} else if lhs.IsKeyPath() && !rhs.IsLiteralArray() {
		return skyerr.NewError(skyerr.InternalQueryInvalid,
			`right operand of "IN" must be an array if comparing with a keypath`)
	}
	return nil
}

func (p Predicate) validateFunctionalPredicate(parentPredicate *Predicate) skyerr.Error {
	expr := p.Children[0].(Expression)
	if expr.Type != Function {
		return skyerr.NewError(skyerr.InternalQueryInvalid,
			`functional predicate must contain functional expression`)
	}

	switch f := expr.Value.(type) {
	case UserRelationFunc:
		if f.RelationName != "_friend" && f.RelationName != "_follow" {
			return skyerr.NewErrorf(skyerr.NotSupported,
				`user relation predicate with "%d" relation is not supported`,
				f.RelationName)
		}
	case UserDiscoverFunc:
		if parentPredicate != nil {
			return skyerr.NewError(skyerr.NotSupported,
				`user discover predicate cannot be combined with other predicates`)
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
	ViewAsUser          *UserInfo
	BypassAccessControl bool
}

// Func is a marker interface to denote a type being a function in skydb.
//
// skydb's function receives zero or more arguments and returns a DataType
// as a result. Result data type is currently omitted in this interface since
// skygear doesn't use it internally yet. In the future it can be utilized to
// provide more extensive type checking at handler level.
type Func interface {
	Args() []interface{}
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

// CountFunc represents a function that count number of rows matching
// a query
type CountFunc struct {
	OverallRecords bool
}

// Args implements the Func interface
func (f CountFunc) Args() []interface{} {
	return []interface{}{}
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

// UserDiscoverFunc searches for user reord having the specified user data, such
// as email addresses. Can only be used with user record.
type UserDiscoverFunc struct {
	Emails []string
}

// Args implements the Func interface
func (f UserDiscoverFunc) Args() []interface{} {
	panic("not supported")
}

// ArgsByName implements the Func interface
func (f UserDiscoverFunc) ArgsByName(name string) []interface{} {
	var data []string
	switch name {
	case "email":
		data = f.Emails
	default:
		panic(fmt.Errorf("not supported arg name %s", name))
	}

	args := make([]interface{}, len(data))
	for i, email := range data {
		args[i] = email
	}
	return args
}

// UserDataFunc is an expresssion to return an attribute of user info
// as email addresses. Can only be used with user record.
type UserDataFunc struct {
	DataName string
}

// Args implements the Func interface
func (f UserDataFunc) Args() []interface{} {
	return []interface{}{}
}
