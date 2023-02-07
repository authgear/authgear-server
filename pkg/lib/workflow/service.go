package workflow

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
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

type Savepoint interface {
	Begin() error
	Rollback() error
	Commit() error
}

type Service struct {
	ContextDoNotUseDirectly context.Context
	Deps                    *Dependencies
	Logger                  ServiceLogger
	Savepoint               Savepoint
	Store                   Store
}

func (s *Service) CreateNewWorkflow(intent Intent, sessionOptions *SessionOptions) (output *ServiceOutput, err error) {
	session := NewSession(sessionOptions)
	err = s.Store.CreateSession(session)
	if err != nil {
		return
	}

	ctx := session.Context(s.ContextDoNotUseDirectly)

	// createNewWorkflow uses defer statement to manage savepoint.
	workflow, workflowOutput, action, err := s.createNewWorkflow(ctx, session, intent)
	// At this point, no savepoint is active.
	if err != nil {
		return
	}

	sessionOutput := session.ToOutput()

	output = &ServiceOutput{
		Session:        session,
		SessionOutput:  sessionOutput,
		Workflow:       workflow,
		WorkflowOutput: workflowOutput,
		Action:         action,
	}

	return
}

func (s *Service) createNewWorkflow(ctx context.Context, session *Session, intent Intent) (workflow *Workflow, output *WorkflowOutput, action *WorkflowAction, err error) {
	// The first thing we need to do is to create a database savepoint.
	err = s.Savepoint.Begin()
	if err != nil {
		return
	}

	// We always rollback.
	defer func() {
		rollbackErr := s.Savepoint.Rollback()
		if rollbackErr != nil {
			if err == nil {
				err = rollbackErr
			} else {
				err = errorutil.WithSecondaryError(err, rollbackErr)
			}
			return
		}
	}()

	// A new workflow does not have any nodes.
	// A workflow is allowed to have on-commit-effects only.
	// So we do not have to apply effects on a new workflow.
	workflow = NewWorkflow(session.WorkflowID, intent)

	err = s.Store.CreateWorkflow(workflow)
	if err != nil {
		return
	}

	output, err = workflow.ToOutput(ctx, s.Deps)
	if err != nil {
		return
	}

	action, err = s.determineAction(ctx, session, workflow)
	if err != nil {
		return
	}

	return
}

func (s *Service) Get(instanceID string) (output *ServiceOutput, err error) {
	w, err := s.Store.GetWorkflowByInstanceID(instanceID)
	if err != nil {
		return
	}

	session, err := s.Store.GetSession(w.WorkflowID)
	if err != nil {
		return
	}

	ctx := session.Context(s.ContextDoNotUseDirectly)

	err = s.Savepoint.Begin()
	if err != nil {
		return
	}

	defer func() {
		rollbackErr := s.Savepoint.Rollback()
		if rollbackErr != nil {
			if err == nil {
				err = rollbackErr
			} else {
				err = errorutil.WithSecondaryError(err, rollbackErr)
			}
			return
		}
	}()

	// Apply the run-effects.
	err = w.ApplyRunEffects(ctx, s.Deps)
	if err != nil {
		return
	}

	workflowOutput, err := w.ToOutput(ctx, s.Deps)
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

func (s *Service) FeedInput(instanceID string, input Input) (output *ServiceOutput, err error) {
	workflow, err := s.Store.GetWorkflowByInstanceID(instanceID)
	if err != nil {
		return
	}

	session, err := s.Store.GetSession(workflow.WorkflowID)
	if err != nil {
		return
	}
	ctx := session.Context(s.ContextDoNotUseDirectly)

	// feedInput uses defer statement to manage savepoint.
	workflow, workflowOutput, action, err := s.feedInput(ctx, session, instanceID, input)
	isEOF := errors.Is(err, ErrEOF)
	// At this point, no savepoint is active.
	if err != nil && !isEOF {
		return
	}

	sessionOutput := session.ToOutput()

	var cookies []*http.Cookie
	if isEOF {
		cookies, err = s.finishWorkflow(ctx, workflow)
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
	// The first thing we need to do is to create a database savepoint.
	err = s.Savepoint.Begin()
	if err != nil {
		return
	}

	// We always rollback.
	defer func() {
		rollbackErr := s.Savepoint.Rollback()
		if rollbackErr != nil {
			if err == nil {
				err = rollbackErr
			} else {
				err = errorutil.WithSecondaryError(err, rollbackErr)
			}
			return
		}
	}()

	workflow, err = s.Store.GetWorkflowByInstanceID(instanceID)
	if err != nil {
		return
	}

	// Apply the run-effects.
	err = workflow.ApplyRunEffects(ctx, s.Deps)
	if err != nil {
		return
	}

	err = workflow.Accept(ctx, s.Deps, input)
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

	output, err = workflow.ToOutput(ctx, s.Deps)
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

	// The first thing we need to do is to create a database savepoint.
	err = s.Savepoint.Begin()
	if err != nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			rollbackErr := s.Savepoint.Rollback()
			if rollbackErr != nil {
				s.Logger.WithError(rollbackErr).Error("workflow failed to rollback")
			}
			panic(r)
		} else if err != nil {
			rollbackErr := s.Savepoint.Rollback()
			if rollbackErr != nil {
				if err == nil {
					err = rollbackErr
				} else {
					err = errorutil.WithSecondaryError(err, rollbackErr)
				}
				return
			}
		} else {
			err = s.Savepoint.Commit()
		}
	}()

	err = workflow.ApplyAllEffects(ctx, s.Deps)
	if err != nil {
		return
	}

	cookies, err = workflow.CollectCookies(ctx, s.Deps)
	if err != nil {
		return
	}

	return
}

func (s *Service) determineAction(ctx context.Context, session *Session, workflow *Workflow) (*WorkflowAction, error) {
	isEOF, err := workflow.IsEOF(ctx, s.Deps)
	if err != nil {
		return nil, err
	}
	if isEOF {
		return &WorkflowAction{
			Type:        WorkflowActionTypeFinish,
			RedirectURI: session.RedirectURI,
		}, nil
	}
	// TODO(workflow): handle oauth redirect.
	return &WorkflowAction{
		Type: WorkflowActionTypeContinue,
	}, nil
}
