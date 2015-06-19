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

// Predicate is an temporary marker struct to denote places where a
// predicate is needed.
type Predicate struct {
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
