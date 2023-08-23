package authenticationflow

type Flow struct {
	FlowID     string
	InstanceID string
	Intent     Intent
	Nodes      []Node
}

func NewFlow(flowID string, intent Intent) *Flow {
	return &Flow{
		FlowID:     flowID,
		InstanceID: newInstanceID(),
		Intent:     intent,
	}
}
