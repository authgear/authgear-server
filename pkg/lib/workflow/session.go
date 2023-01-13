package workflow

type Session struct {
	WorkflowID string

	// Other information we want to keep track of.
	ClientID string
}

type SessionOptions struct {
	ClientID string
}

func NewSession(opts *SessionOptions) *Session {
	return &Session{
		WorkflowID: newWorkflowID(),
		ClientID:   opts.ClientID,
	}
}

func (s *Session) ToOutput() *SessionOutput {
	return &SessionOutput{
		WorkflowID: s.WorkflowID,
		ClientID:   s.ClientID,
	}
}

type SessionOutput struct {
	WorkflowID string `json:"workflow_id"`
	ClientID   string `json:"client_id,omitempty"`
}
