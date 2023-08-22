package workflow

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/log"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package workflow

type WorkflowAction struct {
	Type        WorkflowActionType `json:"type"`
	RedirectURI string             `json:"redirect_uri,omitempty"`
}

type WorkflowActionType string

const (
	WorkflowActionTypeContinue WorkflowActionType = "continue"
	WorkflowActionTypeFinish   WorkflowActionType = "finish"
	WorkflowActionTypeRedirect WorkflowActionType = "redirect"
)

type ServiceOutput struct {
	Session        *Session
	SessionOutput  *SessionOutput
	Workflow       *Workflow
	WorkflowOutput *WorkflowOutput
	Action         *WorkflowAction
	Cookies        []*http.Cookie
}

type ServiceLogger struct{ *log.Logger }

func NewServiceLogger(lf *log.Factory) ServiceLogger {
	return ServiceLogger{lf.New("workflow-service")}
}

type Store interface {
	CreateSession(session *Session) error
	GetSession(workflowID string) (*Session, error)
	DeleteSession(session *Session) error

	CreateWorkflow(workflow *Workflow) error
	GetWorkflowByInstanceID(instanceID string) (*Workflow, error)
	DeleteWorkflow(workflow *Workflow) error
}

type ServiceDatabase interface {
	WithTx(do func() error) (err error)
	ReadOnly(do func() error) (err error)
}

type Service struct {
	ContextDoNotUseDirectly context.Context
	Deps                    *Dependencies
	Logger                  ServiceLogger
	Store                   Store
	Database                ServiceDatabase
}

func (s *Service) CreateNewWorkflow(intent Intent, sessionOptions *SessionOptions) (output *ServiceOutput, err error) {
	session := NewSession(sessionOptions)
	err = s.Store.CreateSession(session)
	if err != nil {
		return
	}

	ctx := session.Context(s.ContextDoNotUseDirectly)

	var workflow *Workflow
	var workflowOutput *WorkflowOutput
	var action *WorkflowAction
	err = s.Database.ReadOnly(func() error {
		workflow, workflowOutput, action, err = s.createNewWorkflow(ctx, session, intent)
		return err
	})
	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	sessionOutput := session.ToOutput()

	var cookies []*http.Cookie
	if isEOF {
		err = s.Database.WithTx(func() error {
			cookies, err = s.finishWorkflow(ctx, workflow)
			return err
		})
		if err != nil {
			return
		}

		err = s.Store.DeleteSession(session)
		if err != nil {
			return
		}

		err = s.Store.DeleteWorkflow(workflow)
		if err != nil {
			return
		}
	}

	if isEOF {
		err = ErrEOF
	}
	output = &ServiceOutput{
		Session:        session,
		SessionOutput:  sessionOutput,
		Workflow:       workflow,
		WorkflowOutput: workflowOutput,
		Action:         action,
		Cookies:        cookies,
	}
	return
}

func (s *Service) createNewWorkflow(ctx context.Context, session *Session, intent Intent) (workflow *Workflow, output *WorkflowOutput, action *WorkflowAction, err error) {
	workflow = NewWorkflow(session.WorkflowID, intent)

	// A new workflow does not have any nodes.
	// A workflow is allowed to have on-commit-effects only.
	// So we do not have to apply effects on a new workflow.

	// Feed an nil input to the workflow to let it proceed.
	var input Input
	err = workflow.Accept(ctx, s.Deps, NewWorkflows(workflow), input)
	// As a special case, we do not treat ErrNoChange as error because
	// Not every workflow can react to nil input.
	if errors.Is(err, ErrNoChange) {
		err = nil
	}
	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	// err is nil or err is ErrEOF.
	// We persist the workflow instance.
	err = s.Store.CreateWorkflow(workflow)
	if err != nil {
		return
	}

	output, err = workflow.ToOutput(ctx, s.Deps, NewWorkflows(workflow))
	if err != nil {
		return
	}

	action, err = s.determineAction(ctx, session, workflow)
	if err != nil {
		return
	}

	if isEOF {
		err = ErrEOF
	}
	return
}

func (s *Service) Get(workflowID string, instanceID string, userAgentID string) (output *ServiceOutput, err error) {
	w, err := s.Store.GetWorkflowByInstanceID(instanceID)
	if err != nil {
		return
	}

	if w.WorkflowID != workflowID {
		err = ErrWorkflowNotFound
		return
	}

	session, err := s.Store.GetSession(w.WorkflowID)
	if err != nil {
		return
	}

	if session.UserAgentID != "" && session.UserAgentID != userAgentID {
		err = ErrUserAgentUnmatched
		return
	}

	ctx := session.Context(s.ContextDoNotUseDirectly)

	err = s.Database.ReadOnly(func() error {
		output, err = s.get(ctx, session, w)
		return err
	})
	return
}

func (s *Service) get(ctx context.Context, session *Session, w *Workflow) (output *ServiceOutput, err error) {
	// Apply the run-effects.
	err = w.ApplyRunEffects(ctx, s.Deps, NewWorkflows(w))
	if err != nil {
		return
	}

	workflowOutput, err := w.ToOutput(ctx, s.Deps, NewWorkflows(w))
	if err != nil {
		return
	}

	action, err := s.determineAction(ctx, session, w)
	if err != nil {
		return
	}

	sessionOutput := session.ToOutput()

	output = &ServiceOutput{
		Session:        session,
		SessionOutput:  sessionOutput,
		Workflow:       w,
		WorkflowOutput: workflowOutput,
		Action:         action,
	}
	return
}

func (s *Service) FeedInput(workflowID string, instanceID string, userAgentID string, input Input) (output *ServiceOutput, err error) {
	workflow, err := s.Store.GetWorkflowByInstanceID(instanceID)
	if err != nil {
		return
	}

	if workflow.WorkflowID != workflowID {
		err = ErrWorkflowNotFound
		return
	}

	session, err := s.Store.GetSession(workflow.WorkflowID)
	if err != nil {
		return
	}

	if session.UserAgentID != "" && session.UserAgentID != userAgentID {
		err = ErrUserAgentUnmatched
		return
	}

	ctx := session.Context(s.ContextDoNotUseDirectly)

	var workflowOutput *WorkflowOutput
	var action *WorkflowAction
	err = s.Database.ReadOnly(func() error {
		workflow, workflowOutput, action, err = s.feedInput(ctx, session, instanceID, input)
		return err
	})
	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	sessionOutput := session.ToOutput()

	var cookies []*http.Cookie
	if isEOF {
		err = s.Database.WithTx(func() error {
			cookies, err = s.finishWorkflow(ctx, workflow)
			return err
		})
		if err != nil {
			return
		}

		err = s.Store.DeleteSession(session)
		if err != nil {
			return
		}

		err = s.Store.DeleteWorkflow(workflow)
		if err != nil {
			return
		}
	}

	if isEOF {
		err = ErrEOF
	}
	output = &ServiceOutput{
		Session:        session,
		SessionOutput:  sessionOutput,
		Workflow:       workflow,
		WorkflowOutput: workflowOutput,
		Action:         action,
		Cookies:        cookies,
	}
	return
}

func (s *Service) feedInput(ctx context.Context, session *Session, instanceID string, input Input) (workflow *Workflow, output *WorkflowOutput, action *WorkflowAction, err error) {
	workflow, err = s.Store.GetWorkflowByInstanceID(instanceID)
	if err != nil {
		return
	}

	// Apply the run-effects.
	err = workflow.ApplyRunEffects(ctx, s.Deps, NewWorkflows(workflow))
	if err != nil {
		return
	}

	err = workflow.Accept(ctx, s.Deps, NewWorkflows(workflow), input)
	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	// err is nil or err is ErrEOF.
	// We persist the workflow instance.
	err = s.Store.CreateWorkflow(workflow)
	if err != nil {
		return
	}

	output, err = workflow.ToOutput(ctx, s.Deps, NewWorkflows(workflow))
	if err != nil {
		return
	}

	action, err = s.determineAction(ctx, session, workflow)
	if err != nil {
		return
	}

	if isEOF {
		err = ErrEOF
	}
	return
}

func (s *Service) finishWorkflow(ctx context.Context, workflow *Workflow) (cookies []*http.Cookie, err error) {
	// When the workflow is finished, we have the following things to do:
	// 1. Apply all effects.
	// 2. Collect cookies.
	err = workflow.ApplyAllEffects(ctx, s.Deps, NewWorkflows(workflow))
	if err != nil {
		return
	}

	cookies, err = workflow.CollectCookies(ctx, s.Deps, NewWorkflows(workflow))
	if err != nil {
		return
	}

	return
}

func (s *Service) determineAction(ctx context.Context, session *Session, workflow *Workflow) (*WorkflowAction, error) {
	isEOF, err := workflow.IsEOF(ctx, s.Deps, NewWorkflows(workflow))
	if err != nil {
		return nil, err
	}
	if isEOF {
		return &WorkflowAction{
			Type:        WorkflowActionTypeFinish,
			RedirectURI: session.RedirectURI,
		}, nil
	}
	return &WorkflowAction{
		Type: WorkflowActionTypeContinue,
	}, nil
}
