package authenticationflow

type Flow struct {
	FlowID  string
	StateID string
	Intent  Intent
	Nodes   []Node
}

func NewFlow(flowID string, intent Intent) *Flow {
	return &Flow{
		FlowID:  flowID,
		StateID: newStateID(),
		Intent:  intent,
	}
}
