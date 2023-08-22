package workflow2

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package workflow2

type ServiceOutput struct {
	Session       *Session
	SessionOutput *SessionOutput
	Workflow      *Workflow
	Data          Data
	SchemaBuilder validation.SchemaBuilder
	Cookies       []*http.Cookie
}

func (o *ServiceOutput) EnsureDataIsNonNil() {
	if o.Data == nil {
		o.Data = EmptyData
	}
}

type determineActionResult struct {
	Data          Data
	SchemaBuilder validation.SchemaBuilder
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
	var determineActionResult *determineActionResult
	err = s.Database.ReadOnly(func() error {
		workflow, determineActionResult, err = s.createNewWorkflow(ctx, session, intent)
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
		Session:       session,
		SessionOutput: sessionOutput,
		Workflow:      workflow,
		Data:          determineActionResult.Data,
		SchemaBuilder: determineActionResult.SchemaBuilder,
		Cookies:       cookies,
	}
	output.EnsureDataIsNonNil()
	return
}

func (s *Service) createNewWorkflow(ctx context.Context, session *Session, intent Intent) (workflow *Workflow, determineActionResult *determineActionResult, err error) {
	workflow = NewWorkflow(session.WorkflowID, intent)

	// A new workflow does not have any nodes.
	// A workflow is allowed to have on-commit-effects only.
	// So we do not have to apply effects on a new workflow.

	// Feed an nil input to the workflow to let it proceed.
	var rawMessage json.RawMessage
	err = Accept(ctx, s.Deps, NewWorkflows(workflow), rawMessage)
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

	determineActionResult, err = s.determineAction(ctx, session, workflow)
	if err != nil {
		return
	}

	if isEOF {
		err = ErrEOF
	}
	return
}

func (s *Service) Get(instanceID string, userAgentID string) (output *ServiceOutput, err error) {
	w, err := s.Store.GetWorkflowByInstanceID(instanceID)
	if err != nil {
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
	err = ApplyRunEffects(ctx, s.Deps, NewWorkflows(w))
	if err != nil {
		return
	}

	determineActionResult, err := s.determineAction(ctx, session, w)
	if err != nil {
		return
	}

	sessionOutput := session.ToOutput()

	output = &ServiceOutput{
		Session:       session,
		SessionOutput: sessionOutput,
		Workflow:      w,
		Data:          determineActionResult.Data,
		SchemaBuilder: determineActionResult.SchemaBuilder,
	}
	output.EnsureDataIsNonNil()
	return
}

func (s *Service) FeedInput(instanceID string, userAgentID string, rawMessage json.RawMessage) (output *ServiceOutput, err error) {
	workflow, err := s.Store.GetWorkflowByInstanceID(instanceID)
	if err != nil {
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

	var determineActionResult *determineActionResult
	err = s.Database.ReadOnly(func() error {
		workflow, determineActionResult, err = s.feedInput(ctx, session, instanceID, rawMessage)
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
		Session:       session,
		SessionOutput: sessionOutput,
		Workflow:      workflow,
		Data:          determineActionResult.Data,
		SchemaBuilder: determineActionResult.SchemaBuilder,
		Cookies:       cookies,
	}
	output.EnsureDataIsNonNil()
	return
}

func (s *Service) feedInput(ctx context.Context, session *Session, instanceID string, rawMessage json.RawMessage) (workflow *Workflow, determineActionResult *determineActionResult, err error) {
	workflow, err = s.Store.GetWorkflowByInstanceID(instanceID)
	if err != nil {
		return
	}

	// Apply the run-effects.
	err = ApplyRunEffects(ctx, s.Deps, NewWorkflows(workflow))
	if err != nil {
		return
	}

	err = Accept(ctx, s.Deps, NewWorkflows(workflow), rawMessage)
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

	determineActionResult, err = s.determineAction(ctx, session, workflow)
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
	err = ApplyAllEffects(ctx, s.Deps, NewWorkflows(workflow))
	if err != nil {
		return
	}

	cookies, err = CollectCookies(ctx, s.Deps, NewWorkflows(workflow))
	if err != nil {
		return
	}

	return
}

func (s *Service) determineAction(ctx context.Context, session *Session, workflow *Workflow) (*determineActionResult, error) {
	findInputReactorResult, err := FindInputReactor(ctx, s.Deps, NewWorkflows(workflow))
	if errors.Is(err, ErrEOF) {
		return &determineActionResult{
			Data: &DataRedirectURI{
				RedirectURI: session.RedirectURI,
			},
		}, nil
	}
	if err != nil {
		return nil, err
	}

	var schemaBuilder validation.SchemaBuilder
	if findInputReactorResult.InputSchema != nil {
		schemaBuilder = findInputReactorResult.InputSchema.SchemaBuilder()
	}

	var data Data
	if dataOutputer, ok := findInputReactorResult.InputReactor.(DataOutputer); ok {
		data, err = dataOutputer.OutputData(ctx, s.Deps, findInputReactorResult.Workflows)
		if err != nil {
			return nil, err
		}
	}

	return &determineActionResult{
		Data:          data,
		SchemaBuilder: schemaBuilder,
	}, nil
}
