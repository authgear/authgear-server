package authenticationflow

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

//go:generate mockgen -source=service.go -destination=service_mock_test.go -package authenticationflow

type ServiceOutput struct {
	Session       *Session
	SessionOutput *SessionOutput
	Flow          *Flow

	FlowReference *FlowReference
	Finished      bool
	FlowAction    *FlowAction

	Cookies []*http.Cookie
}

func (o *ServiceOutput) ToFlowResponse() FlowResponse {
	return FlowResponse{
		StateToken: o.Flow.StateToken,
		Type:       o.FlowReference.Type,
		Name:       o.FlowReference.Name,
		Action:     o.FlowAction,
	}
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
	GetFlowByStateToken(stateToken string) (*Flow, error)
	DeleteFlow(flow *Flow) error
}

type ServiceDatabase interface {
	WithTx(do func() error) (err error)
	ReadOnly(do func() error) (err error)
}

type ServiceUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type OAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

type Service struct {
	ContextDoNotUseDirectly context.Context
	Deps                    *Dependencies
	Logger                  ServiceLogger
	Store                   Store
	Database                ServiceDatabase
	UIConfig                *config.UIConfig
	UIInfoResolver          ServiceUIInfoResolver
	OAuthClientResolver     OAuthClientResolver
}

func (s *Service) CreateNewFlow(publicFlow PublicFlow, sessionOptions *SessionOptions) (output *ServiceOutput, err error) {
	err = s.validateNewFlow(publicFlow, sessionOptions)
	if err != nil {
		return
	}

	session := NewSession(sessionOptions)
	err = s.Store.CreateSession(session)
	if err != nil {
		return
	}

	return s.createNewFlowWithSession(publicFlow, session)
}

func (s *Service) validateNewFlow(publicFlow PublicFlow, sessionOptions *SessionOptions) (err error) {
	// Enforce flow allowlist if clientID is provided.
	if sessionOptions.ClientID != "" {
		flowReference := publicFlow.FlowFlowReference()
		client := s.OAuthClientResolver.ResolveClient(sessionOptions.ClientID)
		if client != nil {
			allowlist := NewFlowAllowlist(client.AuthenticationFlowAllowlist, s.UIConfig.AuthenticationFlow.Groups)
			if !allowlist.CanCreateFlow(flowReference) {
				return ErrFlowNotAllowed
			}
		}
	}

	return err
}

func (s *Service) createNewFlowWithSession(publicFlow PublicFlow, session *Session) (output *ServiceOutput, err error) {
	ctx, err := session.MakeContext(s.ContextDoNotUseDirectly, s.Deps)
	if err != nil {
		return
	}

	var flow *Flow
	var flowAction *FlowAction
	err = s.Database.ReadOnly(func() error {
		flow, flowAction, err = s.createNewFlow(ctx, session, publicFlow)
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

	flowReference := FindCurrentFlowReference(flow)
	output = &ServiceOutput{
		Session:       session,
		SessionOutput: sessionOutput,
		Flow:          flow,
		FlowReference: flowReference,
		FlowAction:    flowAction,
		Cookies:       cookies,
	}
	return
}

func (s *Service) createNewFlow(ctx context.Context, session *Session, publicFlow PublicFlow) (flow *Flow, flowAction *FlowAction, err error) {
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
	// We persist the flow state.
	err = s.Store.CreateFlow(flow)
	if err != nil {
		return
	}

	flowAction, err = s.getFlowAction(ctx, session, flow)
	if err != nil {
		return
	}

	if isEOF {
		err = ErrEOF
	}
	return
}

func (s *Service) Get(stateToken string) (output *ServiceOutput, err error) {
	w, err := s.Store.GetFlowByStateToken(stateToken)
	if err != nil {
		return
	}

	session, err := s.Store.GetSession(w.FlowID)
	if err != nil {
		return
	}

	ctx, err := session.MakeContext(s.ContextDoNotUseDirectly, s.Deps)
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

	flowAction, err := s.getFlowAction(ctx, session, w)
	if err != nil {
		return
	}

	sessionOutput := session.ToOutput()

	flowReference := FindCurrentFlowReference(w)
	output = &ServiceOutput{
		Session:       session,
		SessionOutput: sessionOutput,
		Flow:          w,
		FlowReference: flowReference,
		FlowAction:    flowAction,
	}
	return
}

func (s *Service) FeedInput(stateToken string, rawMessage json.RawMessage) (output *ServiceOutput, err error) {
	if stateToken == "" {
		stateToken, err = s.resolveStateTokenFromInput(rawMessage)
		if err != nil {
			return
		}
	}

	flow, err := s.Store.GetFlowByStateToken(stateToken)
	if err != nil {
		return
	}

	session, err := s.Store.GetSession(flow.FlowID)
	if err != nil {
		return
	}

	ctx, err := session.MakeContext(s.ContextDoNotUseDirectly, s.Deps)
	if err != nil {
		return
	}

	var flowAction *FlowAction
	err = s.Database.ReadOnly(func() error {
		flow, flowAction, err = s.feedInput(ctx, session, stateToken, rawMessage)
		return err
	})

	// Handle switch flow.
	var errSwitchFlow *ErrorSwitchFlow
	if errors.As(err, &errSwitchFlow) {
		output, err = s.switchFlow(session, errSwitchFlow)
		return
	}

	// Handle rewrite flow.
	var errRewriteFlow *ErrorRewriteFlow
	if errors.As(err, &errRewriteFlow) {
		output, err = s.rewriteFlow(session, errRewriteFlow)
		return
	}

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

	flowReference := FindCurrentFlowReference(flow)
	output = &ServiceOutput{
		Session:       session,
		SessionOutput: sessionOutput,
		Flow:          flow,
		FlowReference: flowReference,
		FlowAction:    flowAction,
		Cookies:       cookies,
	}
	return
}

func (s *Service) FeedSyntheticInput(stateToken string, syntheticInput Input) (output *ServiceOutput, err error) {
	flow, err := s.Store.GetFlowByStateToken(stateToken)
	if err != nil {
		return
	}

	session, err := s.Store.GetSession(flow.FlowID)
	if err != nil {
		return
	}

	ctx, err := session.MakeContext(s.ContextDoNotUseDirectly, s.Deps)
	if err != nil {
		return
	}

	var flowAction *FlowAction
	err = s.Database.ReadOnly(func() error {
		flow, flowAction, err = s.feedSyntheticInput(ctx, session, stateToken, syntheticInput)
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

	flowReference := FindCurrentFlowReference(flow)
	output = &ServiceOutput{
		Session:       session,
		SessionOutput: sessionOutput,
		Flow:          flow,
		FlowReference: flowReference,
		FlowAction:    flowAction,
		Cookies:       cookies,
	}
	return
}

func (s *Service) switchFlow(session *Session, errSwitchFlow *ErrorSwitchFlow) (output *ServiceOutput, err error) {
	publicFlow, err := InstantiateFlow(errSwitchFlow.FlowReference, jsonpointer.T{})
	if err != nil {
		return
	}

	createOutput, err := s.createNewFlowWithSession(publicFlow, session)
	if err != nil {
		return
	}

	output, err = s.FeedSyntheticInput(createOutput.Flow.StateToken, errSwitchFlow.SyntheticInput)
	if err != nil {
		return
	}

	var cookies []*http.Cookie
	for _, c := range createOutput.Cookies {
		cookies = append(cookies, c)
	}
	for _, c := range output.Cookies {
		cookies = append(cookies, c)
	}
	output.Cookies = cookies

	return
}

func (s *Service) rewriteFlow(session *Session, errRewriteFlow *ErrorRewriteFlow) (output *ServiceOutput, err error) {
	newFlow := NewFlow(session.FlowID, errRewriteFlow.Intent)
	newFlow.Nodes = errRewriteFlow.Nodes
	err = s.Store.CreateFlow(newFlow)
	if err != nil {
		return
	}
	return s.FeedSyntheticInput(newFlow.StateToken, errRewriteFlow.SyntheticInput)
}

func (s *Service) feedInput(ctx context.Context, session *Session, stateToken string, rawMessage json.RawMessage) (flow *Flow, flowAction *FlowAction, err error) {
	flow, err = s.Store.GetFlowByStateToken(stateToken)
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
	// We persist the flow state.
	err = s.Store.CreateFlow(flow)
	if err != nil {
		return
	}

	flowAction, err = s.getFlowAction(ctx, session, flow)
	if err != nil {
		return
	}

	if isEOF {
		err = ErrEOF
	}
	return
}

func (s *Service) feedSyntheticInput(ctx context.Context, session *Session, stateToken string, syntheticInput Input) (flow *Flow, flowAction *FlowAction, err error) {
	flow, err = s.Store.GetFlowByStateToken(stateToken)
	if err != nil {
		return
	}

	// Apply the run-effects.
	err = ApplyRunEffects(ctx, s.Deps, NewFlows(flow))
	if err != nil {
		return
	}

	err = AcceptSyntheticInput(ctx, s.Deps, NewFlows(flow), syntheticInput)
	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	// err is nil or err is ErrEOF.
	// We persist the flow state.
	err = s.Store.CreateFlow(flow)
	if err != nil {
		return
	}

	flowAction, err = s.getFlowAction(ctx, session, flow)
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

func (s *Service) getFlowAction(ctx context.Context, session *Session, flow *Flow) (flowAction *FlowAction, err error) {
	findInputReactorResult, err := FindInputReactor(ctx, s.Deps, NewFlows(flow))
	if errors.Is(err, ErrEOF) {
		redirectURI := session.RedirectURI
		e, ok := GetAuthenticationInfoEntry(ctx, s.Deps, NewFlows(flow))
		if ok {
			redirectURI = s.UIInfoResolver.SetAuthenticationInfoInQuery(redirectURI, e)
		}

		dataFinishRedirectURI := &DataFinishRedirectURI{
			FinishRedirectURI: redirectURI,
		}
		var data Data = dataFinishRedirectURI

		if outputer, ok := flow.Intent.(EndOfFlowDataOutputer); ok {
			data, err = outputer.OutputEndOfFlowData(ctx, s.Deps, NewFlows(flow), dataFinishRedirectURI)
			if err != nil {
				return nil, err
			}
		}

		flowAction = &FlowAction{
			Type: FlowActionTypeFinished,
			Data: data,
		}
		return
	}
	if err != nil {
		return nil, err
	}

	if findInputReactorResult.InputSchema != nil {
		p := findInputReactorResult.InputSchema.GetJSONPointer()
		flowRootObject := findInputReactorResult.InputSchema.GetFlowRootObject()

		if flowRootObject != nil {
			flowAction = GetFlowAction(flowRootObject, p)
		}
	}

	var data Data
	if dataOutputer, ok := findInputReactorResult.InputReactor.(DataOutputer); ok {
		data, err = dataOutputer.OutputData(ctx, s.Deps, findInputReactorResult.Flows)
		if err != nil {
			return nil, err
		}
	}
	if data == nil {
		data = mapData{}
	}
	if flowAction != nil {
		flowAction.Data = data
	}

	return
}

func (s *Service) resolveStateTokenFromInput(inputRawMessage json.RawMessage) (string, error) {
	if input, ok := MakeInputTakeAccountRecoveryCode(inputRawMessage); ok {
		state, err := s.Deps.ResetPassword.VerifyCode(input.AccountRecoveryCode)
		if err != nil {
			return "", err
		}
		flow, err := InstantiateFlow(FlowReference{
			Type: FlowType(state.AuthenticationFlowType),
			Name: state.AuthenticationFlowName,
		}, state.AuthenticationFlowJSONPointer)
		if err != nil {
			return "", err
		}
		// In account recovery flow, session options are not important
		newFlowOutput, err := s.CreateNewFlow(flow, &SessionOptions{})
		if err != nil {
			return "", err
		}
		return newFlowOutput.Flow.StateToken, nil

	}
	return "", ErrFlowNotFound
}
