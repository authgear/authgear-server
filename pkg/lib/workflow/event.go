package workflow

type EventKind string

const (
	// WorkflowEventKindRefresh indicates client should re-fetch current instance of workflow for updated state.
	EventKindRefresh EventKind = "refresh"
)

type Event interface {
	kind() EventKind
}

type EventRefresh struct {
	Kind EventKind `json:"kind"`
}

func NewEventRefresh() *EventRefresh {
	return &EventRefresh{Kind: EventKindRefresh}
}

func (*EventRefresh) kind() EventKind { return EventKindRefresh }

var _ Event = &EventRefresh{}
