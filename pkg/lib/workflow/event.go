package workflow

type EventKind string

const (
	// WorkflowEventKindRefresh indicates client should re-fetch current instance of workflow for updated state.
	EventKindRefresh EventKind = "refresh"
	// WorkflowEventKindLoginLinkCodeVerified indicates client should proceed since login link code is verified by server.
	WorkflowEventKindLoginLinkCodeVerified EventKind = "login-link-code-verified"
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

type EventLoginLinkCodeVerified struct {
	Kind EventKind `json:"kind"`
}

func NewEventLoginLinkCodeVerified() *EventLoginLinkCodeVerified {
	return &EventLoginLinkCodeVerified{Kind: WorkflowEventKindLoginLinkCodeVerified}
}

func (*EventLoginLinkCodeVerified) kind() EventKind { return WorkflowEventKindLoginLinkCodeVerified }

var _ Event = &EventLoginLinkCodeVerified{}
