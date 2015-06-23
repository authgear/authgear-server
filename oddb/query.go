package oddb

// SortOrder denotes an the order of Records returned from a Query.
type SortOrder int

// A list of SordOrder, their meaning is self descriptive.
const (
	Ascending SortOrder = iota
	Descending
	Asc  = Ascending
	Desc = Descending
)

// Sort specifies the field and the order to be sorted against a collection of
// Records returned from a Query.
type Sort struct {
	KeyPath string
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
)

// Checks whether the Operator is a compound operator, meaning the
// operator combine the results of other subpredicates.
func (op Operator) IsCompound() bool {
	switch op {
	default:
		return false
	case And, Or, Not:
		return true
	}
}

// Checks whether the Operator determines the result of a predicate
// by comparing two subexpressions.
func (op Operator) IsBinary() bool {
	switch op {
	default:
		return false
	case Equal, GreaterThan, LessThan, GreaterThanOrEqual, LessThanOrEqual, NotEqual:
		return true
	}
}

// Returns the type of an Expression.
type ExpressionType int

// A list of ExpressionTypes.
const (
	Literal ExpressionType = iota
	KeyPath
)

// An Expression represents value to be compared against.
type Expression struct {
	Type  ExpressionType
	Value interface{}
}

// Predicate is a representation of used in query for filtering records.
type Predicate struct {
	Operator Operator
	Children []interface{}
}

// Query specifies the type, predicate and sorting order of Database
// query.
// ReadableBy is a temp solution for ACL before a full predicate implemented.
type Query struct {
	Type       string     `json:"record_type"`
	Predicate  *Predicate `json:"predicate,omitempty"`
	Sorts      []Sort     `json:"order,omitempty"`
	ReadableBy string     `json:"readable_by,omitempty"`
}
