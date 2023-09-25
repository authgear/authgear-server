package authenticationflow

type Flow struct {
	FlowID     string
	StateToken string
	Intent     Intent
	Nodes      []Node
}

func NewFlow(flowID string, intent Intent) *Flow {
	return &Flow{
		FlowID:     flowID,
		StateToken: newStateToken(),
		Intent:     intent,
	}
}
