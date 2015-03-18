package oddb

// RecordHookEvent indicates the type of record event that triggered
// the hook
type RecordHookEvent int

// See the definition of RecordHookEvent
const (
	RecordCreated RecordHookEvent = iota + 1
	RecordUpdated
	RecordDeleted
)
