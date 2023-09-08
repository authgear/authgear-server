package authenticationflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package authenticationflow

type ServiceOutput struct {
	Session       *Session
	SessionOutput *SessionOutput
	Flow          *Flow

	Finished      bool
	SchemaBuilder validation.SchemaBuilder

	FlowReference *FlowReference
	FlowStep      *FlowStep
	Data          Data

	Cookies []*http.Cookie
}

func (o *ServiceOutput) ToFlowResponse() FlowResponse {
	return FlowResponse{
		ID:            o.Flow.InstanceID,
		WebsocketID:   o.Flow.FlowID,
		JSONSchema:    o.SchemaBuilder,
		Finished:      o.Finished,
		FlowReference: o.FlowReference,
		FlowStep:      o.FlowStep,
		Data:          o.Data,
	}
}

type determineActionResult struct {
	Finished      bool
	FlowReference *FlowReference
	FlowStep      *FlowStep
	Data          Data
	SchemaBuilder validation.SchemaBuilder
}

type ServiceLogger struct{ *log.Logger }

func NewServiceLogger(lf *log.Factory) ServiceLogger {
	return ServiceLogger{lf.New("authenticationflow-service")}
}

type Store interface {
	CreateSession(session *Session) error
	GetSession(flowID string) (*Session, error)
	DeleteSession(session *Session) error

	CreateFlow(flow *Flow) error
	GetFlowByInstanceID(instanceID string) (*Flow, error)
	DeleteFlow(flow *Flow) error
}

type ServiceDatabase interface {
	WithTx(do func() error) (err error)
	ReadOnly(do func() error) (err error)
}

type ServiceUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type Service struct {
	ContextDoNotUseDirectly context.Context
	Deps                    *Dependencies
	Logger                  ServiceLogger
	Store                   Store
	Database                ServiceDatabase
	UIInfoResolver          ServiceUIInfoResolver
}

func (s *Service) CreateNewFlow(publicFlow PublicFlow, sessionOptions *SessionOptions) (output *ServiceOutput, err error) {
	session := NewSession(sessionOptions)
	err = s.Store.CreateSession(session)
	if err != nil {
		return
	}

	ctx, err := session.MakeContext(s.ContextDoNotUseDirectly, s.Deps, publicFlow)
	if err != nil {
		return
	}

	var flow *Flow
	var determineActionResult *determineActionResult
	err = s.Database.ReadOnly(func() error {
		flow, determineActionResult, err = s.createNewFlow(ctx, session, publicFlow)
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
			cookies, err = s.finishFlow(ctx, flow)
			return err
		})
		if err != nil {
			return
		}

		err = s.Store.DeleteSession(session)
		if err != nil {
			return
		}

		err = s.Store.DeleteFlow(flow)
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
		Flow:          flow,
		FlowReference: determineActionResult.FlowReference,
		FlowStep:      determineActionResult.FlowStep,
		Data:          determineActionResult.Data,
		Finished:      determineActionResult.Finished,
		SchemaBuilder: determineActionResult.SchemaBuilder,
		Cookies:       cookies,
	}
	return
}

func (s *Service) createNewFlow(ctx context.Context, session *Session, publicFlow PublicFlow) (flow *Flow, determineActionResult *determineActionResult, err error) {
	flow = NewFlow(session.FlowID, publicFlow)

	// A new flow does not have any nodes.
	// A flow is allowed to have on-commit-effects only.
	// So we do not have to apply effects on a new flow.

	// Feed an nil input to the flow to let it proceed.
	var rawMessage json.RawMessage
	err = Accept(ctx, s.Deps, NewFlows(flow), rawMessage)
	// As a special case, we do not treat ErrNoChange as error because
	// Not every flow can react to nil input.
	if errors.Is(err, ErrNoChange) {
		err = nil
	}
	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	// err is nil or err is ErrEOF.
	// We persist the flow instance.
	err = s.Store.CreateFlow(flow)
	if err != nil {
		return
	}

	determineActionResult, err = s.determineAction(ctx, session, flow)
	if err != nil {
		return
	}

	if isEOF {
		err = ErrEOF
	}
	return
}

func (s *Service) Get(instanceID string, userAgentID string) (output *ServiceOutput, err error) {
	w, err := s.Store.GetFlowByInstanceID(instanceID)
	if err != nil {
		return
	}

	session, err := s.Store.GetSession(w.FlowID)
	if err != nil {
		return
	}

	if session.UserAgentID != "" && session.UserAgentID != userAgentID {
		err = ErrUserAgentUnmatched
		return
	}

	publicFlow, ok := w.Intent.(PublicFlow)
	if !ok {
		panic(fmt.Errorf("the root intent must be a PublicFlow: %T", w.Intent))
	}

	ctx, err := session.MakeContext(s.ContextDoNotUseDirectly, s.Deps, publicFlow)
	if err != nil {
		return
	}

	err = s.Database.ReadOnly(func() error {
		output, err = s.get(ctx, session, w)
		return err
	})
	return
}

func (s *Service) get(ctx context.Context, session *Session, w *Flow) (output *ServiceOutput, err error) {
	// Apply the run-effects.
	err = ApplyRunEffects(ctx, s.Deps, NewFlows(w))
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
		Flow:          w,
		FlowReference: determineActionResult.FlowReference,
		FlowStep:      determineActionResult.FlowStep,
		Data:          determineActionResult.Data,
		SchemaBuilder: determineActionResult.SchemaBuilder,
		Finished:      determineActionResult.Finished,
	}
	return
}

func (s *Service) FeedInput(instanceID string, userAgentID string, rawMessage json.RawMessage) (output *ServiceOutput, err error) {
	flow, err := s.Store.GetFlowByInstanceID(instanceID)
	if err != nil {
		return
	}

	session, err := s.Store.GetSession(flow.FlowID)
	if err != nil {
		return
	}

	if session.UserAgentID != "" && session.UserAgentID != userAgentID {
		err = ErrUserAgentUnmatched
		return
	}

	publicFlow, ok := flow.Intent.(PublicFlow)
	if !ok {
		panic(fmt.Errorf("the root intent must be a PublicFlow: %T", flow.Intent))
	}

	ctx, err := session.MakeContext(s.ContextDoNotUseDirectly, s.Deps, publicFlow)
	if err != nil {
		return
	}

	var determineActionResult *determineActionResult
	err = s.Database.ReadOnly(func() error {
		flow, determineActionResult, err = s.feedInput(ctx, session, instanceID, rawMessage)
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
			cookies, err = s.finishFlow(ctx, flow)
			return err
		})
		if err != nil {
			return
		}

		err = s.Store.DeleteSession(session)
		if err != nil {
			return
		}

		err = s.Store.DeleteFlow(flow)
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
		Flow:          flow,
		FlowReference: determineActionResult.FlowReference,
		FlowStep:      determineActionResult.FlowStep,
		Data:          determineActionResult.Data,
		SchemaBuilder: determineActionResult.SchemaBuilder,
		Finished:      determineActionResult.Finished,
		Cookies:       cookies,
	}
	return
}

func (s *Service) feedInput(ctx context.Context, session *Session, instanceID string, rawMessage json.RawMessage) (flow *Flow, determineActionResult *determineActionResult, err error) {
	flow, err = s.Store.GetFlowByInstanceID(instanceID)
	if err != nil {
		return
	}

	// Apply the run-effects.
	err = ApplyRunEffects(ctx, s.Deps, NewFlows(flow))
	if err != nil {
		return
	}

	err = Accept(ctx, s.Deps, NewFlows(flow), rawMessage)
	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	// err is nil or err is ErrEOF.
	// We persist the flow instance.
	err = s.Store.CreateFlow(flow)
	if err != nil {
		return
	}

	determineActionResult, err = s.determineAction(ctx, session, flow)
	if err != nil {
		return
	}

	if isEOF {
		err = ErrEOF
	}
	return
}

func (s *Service) finishFlow(ctx context.Context, flow *Flow) (cookies []*http.Cookie, err error) {
	// When the flow is finished, we have the following things to do:
	// 1. Apply all effects.
	// 2. Collect cookies.
	err = ApplyAllEffects(ctx, s.Deps, NewFlows(flow))
	if err != nil {
		return
	}

	cookies, err = CollectCookies(ctx, s.Deps, NewFlows(flow))
	if err != nil {
		return
	}

	return
}

func (s *Service) determineAction(ctx context.Context, session *Session, flow *Flow) (result *determineActionResult, err error) {
	flowReference := GetFlowReference(ctx)

	findInputReactorResult, err := FindInputReactor(ctx, s.Deps, NewFlows(flow))
	if errors.Is(err, ErrEOF) {
		redirectURI := session.RedirectURI
		e, ok := GetAuthenticationInfoEntry(ctx, s.Deps, NewFlows(flow))
		if ok {
			redirectURI = s.UIInfoResolver.SetAuthenticationInfoInQuery(redirectURI, e)
		}
		result = &determineActionResult{
			Finished:      true,
			FlowReference: &flowReference,
			Data: &DataFinishRedirectURI{
				FinishRedirectURI: redirectURI,
			},
		}
		return
	}
	if err != nil {
		return nil, err
	}

	var flowStep *FlowStep
	var schemaBuilder validation.SchemaBuilder
	if findInputReactorResult.InputSchema != nil {
		schemaBuilder = findInputReactorResult.InputSchema.SchemaBuilder()
		p := findInputReactorResult.InputSchema.GetJSONPointer()
		flowRootObject := GetFlowRootObject(ctx)
		if flowRootObject != nil {
			flowStep = GetFlowStep(flowRootObject, p)
		}
	}

	// Ensure data is always non-nil.
	var data Data = mapData{}
	if dataOutputer, ok := findInputReactorResult.InputReactor.(DataOutputer); ok {
		data, err = dataOutputer.OutputData(ctx, s.Deps, findInputReactorResult.Flows)
		if err != nil {
			return nil, err
		}
	}

	result = &determineActionResult{
		FlowReference: &flowReference,
		FlowStep:      flowStep,
		Data:          data,
		SchemaBuilder: schemaBuilder,
	}
	return
}
