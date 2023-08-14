package workflow2

type Workflow struct {
	WorkflowID string
	InstanceID string
	Intent     Intent
	Nodes      []Node
}

func NewWorkflow(workflowID string, intent Intent) *Workflow {
	return &Workflow{
		WorkflowID: workflowID,
		InstanceID: newInstanceID(),
		Intent:     intent,
	}
}
