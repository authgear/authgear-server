package workflow

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package workflow

type ServiceOutput struct {
	Session        *Session
	SessionOutput  *SessionOutput
	Workflow       *Workflow
	WorkflowOutput *WorkflowOutput
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
	workflow, workflowOutput, err := s.createNewWorkflow(ctx, session.WorkflowID, intent)
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
	}

	return
}

func (s *Service) createNewWorkflow(ctx context.Context, workflowID string, intent Intent) (workflow *Workflow, output *WorkflowOutput, err error) {
	// The first thing we need to do is to create a database savepoint.
	err = s.Savepoint.Begin()

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
	workflow = NewWorkflow(workflowID, intent)

	err = s.Store.CreateWorkflow(workflow)
	if err != nil {
		return
	}

	output, err = workflow.ToOutput(ctx, s.Deps)
	return
}

func (s *Service) FeedInput(workflowID string, instanceID string, input interface{}) (output *ServiceOutput, err error) {
	session, err := s.Store.GetSession(workflowID)
	if err != nil {
		return
	}
	ctx := session.Context(s.ContextDoNotUseDirectly)

	// feedInput uses defer statement to manage savepoint.
	workflow, workflowOutput, err := s.feedInput(ctx, instanceID, input)
	isEOF := errors.Is(err, ErrEOF)
	// At this point, no savepoint is active.
	if err != nil && !isEOF {
		return
	}

	sessionOutput := session.ToOutput()

	if isEOF {
		// When the workflow is finished.
		// We need to apply the all effects.
		err = s.applyFinishedWorkflow(ctx, workflow)
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
	}
	return
}

func (s *Service) feedInput(ctx context.Context, instanceID string, input interface{}) (workflow *Workflow, output *WorkflowOutput, err error) {
	// The first thing we need to do is to create a database savepoint.
	err = s.Savepoint.Begin()

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

	if isEOF {
		err = ErrEOF
	}
	return
}

func (s *Service) applyFinishedWorkflow(ctx context.Context, workflow *Workflow) (err error) {
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

	return
}
