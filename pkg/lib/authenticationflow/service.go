package authenticationflow

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

//go:generate go tool mockgen -source=service.go -destination=service_mock_test.go -package authenticationflow

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

var ServiceLogger = slogutil.NewLogger("authenticationflow-service")

type Store interface {
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, flowID string) (*Session, error)
	DeleteSession(ctx context.Context, session *Session) error
	UpdateSession(ctx context.Context, session *Session) error

	CreateFlow(ctx context.Context, flow *Flow) error
	GetFlowByStateToken(ctx context.Context, stateToken string) (*Flow, error)
	DeleteFlow(ctx context.Context, flow *Flow) error
}

type ServiceDatabase interface {
	WithTx(ctx context.Context, do func(ctx context.Context) error) (err error)
	ReadOnly(ctx context.Context, do func(ctx context.Context) error) (err error)
}

type ServiceUIInfoResolver interface {
	SetAuthenticationInfoInQuery(redirectURI string, e *authenticationinfo.Entry) string
}

type OAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

type OAuthSessionStore interface {
	Get(ctx context.Context, entryID string) (entry *oauthsession.Entry, err error)
	Save(ctx context.Context, entry *oauthsession.Entry) (err error)
}

type Service struct {
	Deps                *Dependencies
	Store               Store
	Database            ServiceDatabase
	UIConfig            *config.UIConfig
	UIInfoResolver      ServiceUIInfoResolver
	OAuthClientResolver OAuthClientResolver
	OAuthSessionStore   OAuthSessionStore
}

func (s *Service) CreateNewFlow(ctx context.Context, publicFlow PublicFlow, sessionOptions *SessionOptions) (output *ServiceOutput, err error) {
	err = s.validateNewFlow(publicFlow, sessionOptions)
	if err != nil {
		return
	}

	session := NewSession(sessionOptions)
	ctx = session.MakeContext(ctx, s.Deps)

	err = s.Store.CreateSession(ctx, session)
	if err != nil {
		return
	}

	otelutil.IntCounterAddOne(
		ctx,
		otelauthgear.CounterAuthflowSessionCreationCount,
	)

	return s.createNewFlowWithSession(ctx, publicFlow, session)
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

func (s *Service) createNewFlowWithSession(ctx context.Context, publicFlow PublicFlow, session *Session) (output *ServiceOutput, err error) {
	var flow *Flow
	var flowAction *FlowAction
	flow, flowAction, err = s.createNewFlow(ctx, session, publicFlow)

	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	sessionOutput := session.ToOutput()

	var cookies []*http.Cookie
	if isEOF {
		err = s.Database.WithTx(ctx, func(ctx context.Context) error {
			cookies, err = s.finishFlow(ctx, flow)
			return err
		})
		if err != nil {
			return
		}

		err = s.Store.DeleteSession(ctx, session)
		if err != nil {
			return
		}

		err = s.Store.DeleteFlow(ctx, flow)
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

func (s *Service) processAcceptResult(
	ctx context.Context,
	session *Session,
	flows Flows,
	acceptResult *AcceptResult,
) error {
	if acceptResult.BotProtectionVerificationResult != nil {
		session.SetBotProtectionVerificationResult(acceptResult.BotProtectionVerificationResult)
		updateSessionErr := s.Store.UpdateSession(ctx, session)
		if updateSessionErr != nil {
			return updateSessionErr
		}
	}
	for _, fn := range acceptResult.DelayedOneTimeFunctions {
		err := fn(ctx, s.Deps)
		if err != nil {
			err = s.Database.ReadOnly(ctx, func(ctx context.Context) error {
				// Restore the database state
				runEffectErr := ApplyRunEffects(ctx, s.Deps, flows)
				if runEffectErr != nil {
					return errors.Join(runEffectErr, err)
				}
				err = logAuthenticationBlockedErrorIfNeeded(ctx, s.Deps, flows, err)
				return newAuthenticationFlowError(flows, err)
			})
			return err
		}
	}
	return nil
}

func (s *Service) createNewFlow(ctx context.Context, session *Session, publicFlow PublicFlow) (flow *Flow, flowAction *FlowAction, err error) {
	flow = NewFlow(session.FlowID, publicFlow)

	// A new flow does not have any nodes.
	// A flow is allowed to have on-commit-effects only.
	// So we do not have to apply effects on a new flow.

	// Feed an nil input to the flow to let it proceed.
	var rawMessage json.RawMessage
	var shouldAccept = true
	for shouldAccept {
		shouldAccept = false
		flows := NewFlows(flow)
		var acceptResult *AcceptResult = NewAcceptResult()
		err = s.Database.ReadOnly(ctx, func(ctx context.Context) error {
			err = Accept(ctx, s.Deps, flows, acceptResult, rawMessage)
			isEOF := errors.Is(err, ErrEOF)
			if err != nil && !isEOF {
				return err
			}
			flowAction, err = s.getFlowAction(ctx, session, flow)
			if err != nil {
				return err
			}
			if isEOF {
				return ErrEOF
			}
			return nil
		})
		acceptErr := s.processAcceptResult(ctx, session, flows, acceptResult)
		if acceptErr != nil {
			return nil, nil, acceptErr
		}
		if errors.Is(err, ErrPauseAndRetryAccept) {
			shouldAccept = true
			err = nil
		}
	}

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
	err = s.Store.CreateFlow(ctx, flow)
	if err != nil {
		return
	}

	if isEOF {
		err = ErrEOF
	}
	return
}

func (s *Service) Get(ctx context.Context, stateToken string) (output *ServiceOutput, err error) {
	w, err := s.Store.GetFlowByStateToken(ctx, stateToken)
	if err != nil {
		return
	}

	ctx, session, err := s.getSessionAndUpdateContext(ctx, w.FlowID)
	if err != nil {
		return
	}

	err = s.Database.ReadOnly(ctx, func(ctx context.Context) error {
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

func (s *Service) FeedInput(ctx context.Context, stateToken string, rawMessage json.RawMessage) (output *ServiceOutput, err error) {
	if stateToken == "" {
		stateToken, err = s.resolveStateTokenFromInput(ctx, rawMessage)
		if err != nil {
			return
		}
	}

	flow, err := s.Store.GetFlowByStateToken(ctx, stateToken)
	if err != nil {
		return
	}

	ctx, session, err := s.getSessionAndUpdateContext(ctx, flow.FlowID)
	if err != nil {
		return
	}

	var flowAction *FlowAction
	flow, flowAction, err = s.feedInput(ctx, session, stateToken, rawMessage)

	var errSwitchFlow *ErrorSwitchFlow
	var errRewriteFlow *ErrorRewriteFlow
	isSpecialError := false
	for errors.As(err, &errSwitchFlow) || errors.As(err, &errRewriteFlow) {
		isSpecialError = true
		if errors.As(err, &errSwitchFlow) {
			output, err = s.switchFlow(ctx, session, errSwitchFlow)
		}

		if errors.As(err, &errRewriteFlow) {
			output, err = s.rewriteFlow(ctx, session, errRewriteFlow)
		}
	}

	if isSpecialError {
		return
	}

	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	sessionOutput := session.ToOutput()

	var cookies []*http.Cookie
	if isEOF {
		err = s.Database.WithTx(ctx, func(ctx context.Context) error {
			cookies, err = s.finishFlow(ctx, flow)
			return err
		})
		if err != nil {
			return
		}

		err = s.Store.DeleteSession(ctx, session)
		if err != nil {
			return
		}

		err = s.Store.DeleteFlow(ctx, flow)
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

func (s *Service) FeedSyntheticInput(ctx context.Context, stateToken string, syntheticInput Input) (output *ServiceOutput, err error) {
	flow, err := s.Store.GetFlowByStateToken(ctx, stateToken)
	if err != nil {
		return
	}

	ctx, session, err := s.getSessionAndUpdateContext(ctx, flow.FlowID)
	if err != nil {
		return
	}

	var flowAction *FlowAction
	flow, flowAction, err = s.feedSyntheticInput(ctx, session, stateToken, syntheticInput)

	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	sessionOutput := session.ToOutput()

	var cookies []*http.Cookie
	if isEOF {
		err = s.Database.WithTx(ctx, func(ctx context.Context) error {
			cookies, err = s.finishFlow(ctx, flow)
			return err
		})
		if err != nil {
			return
		}

		err = s.Store.DeleteSession(ctx, session)
		if err != nil {
			return
		}

		err = s.Store.DeleteFlow(ctx, flow)
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

func (s *Service) switchFlow(ctx context.Context, session *Session, errSwitchFlow *ErrorSwitchFlow) (output *ServiceOutput, err error) {
	publicFlow, err := InstantiateFlow(errSwitchFlow.FlowReference, jsonpointer.T{})
	if err != nil {
		return
	}

	createOutput, err := s.createNewFlowWithSession(ctx, publicFlow, session)
	if err != nil {
		return
	}

	output, err = s.FeedSyntheticInput(ctx, createOutput.Flow.StateToken, errSwitchFlow.SyntheticInput)
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

func (s *Service) rewriteFlow(ctx context.Context, session *Session, errRewriteFlow *ErrorRewriteFlow) (output *ServiceOutput, err error) {
	newFlow := NewFlow(session.FlowID, errRewriteFlow.Intent)
	newFlow.Nodes = errRewriteFlow.Nodes
	err = s.Store.CreateFlow(ctx, newFlow)
	if err != nil {
		return
	}
	return s.FeedSyntheticInput(ctx, newFlow.StateToken, errRewriteFlow.SyntheticInput)
}

func (s *Service) feedInput(ctx context.Context, session *Session, stateToken string, rawMessage json.RawMessage) (flow *Flow, flowAction *FlowAction, err error) {
	flow, err = s.Store.GetFlowByStateToken(ctx, stateToken)
	if err != nil {
		return
	}

	var shouldAccept = true
	for shouldAccept {
		shouldAccept = false
		var acceptResult *AcceptResult = NewAcceptResult()
		flows := NewFlows(flow)
		err = s.Database.ReadOnly(ctx, func(ctx context.Context) error {
			// Apply the run-effects.
			err = ApplyRunEffects(ctx, s.Deps, flows)
			if err != nil {
				return err
			}

			err = Accept(ctx, s.Deps, flows, acceptResult, rawMessage)
			isEOF := errors.Is(err, ErrEOF)
			if err != nil && !isEOF {
				return err
			}
			flowAction, err = s.getFlowAction(ctx, session, flow)
			if err != nil {
				return err
			}
			if isEOF {
				return ErrEOF
			}
			return nil
		})
		acceptErr := s.processAcceptResult(ctx, session, flows, acceptResult)
		if acceptErr != nil {
			return nil, nil, acceptErr
		}

		if errors.Is(err, ErrPauseAndRetryAccept) {
			shouldAccept = true
			err = nil
		}
	}

	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	// err is nil or err is ErrEOF.
	// We persist the flow state.
	err = s.Store.CreateFlow(ctx, flow)
	if err != nil {
		return
	}

	if isEOF {
		err = ErrEOF
	}
	return
}

func (s *Service) feedSyntheticInput(ctx context.Context, session *Session, stateToken string, syntheticInput Input) (flow *Flow, flowAction *FlowAction, err error) {
	flow, err = s.Store.GetFlowByStateToken(ctx, stateToken)
	if err != nil {
		return
	}

	var shouldAccept = true
	for shouldAccept {
		shouldAccept = false
		var acceptResult *AcceptResult = NewAcceptResult()
		flows := NewFlows(flow)
		err = s.Database.ReadOnly(ctx, func(ctx context.Context) error {
			// Apply the run-effects.
			err = ApplyRunEffects(ctx, s.Deps, flows)
			if err != nil {
				return err
			}

			err = AcceptSyntheticInput(ctx, s.Deps, flows, acceptResult, syntheticInput)
			isEOF := errors.Is(err, ErrEOF)
			if err != nil && !isEOF {
				return err
			}
			flowAction, err = s.getFlowAction(ctx, session, flow)
			if err != nil {
				return err
			}
			if isEOF {
				return ErrEOF
			}
			return nil
		})
		acceptErr := s.processAcceptResult(ctx, session, flows, acceptResult)
		if acceptErr != nil {
			return nil, nil, acceptErr
		}

		if errors.Is(err, ErrPauseAndRetryAccept) {
			shouldAccept = true
			err = nil
		}
	}

	isEOF := errors.Is(err, ErrEOF)
	if err != nil && !isEOF {
		return
	}

	// err is nil or err is ErrEOF.
	// We persist the flow state.
	err = s.Store.CreateFlow(ctx, flow)
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

func (s *Service) resolveStateTokenFromInput(ctx context.Context, inputRawMessage json.RawMessage) (string, error) {
	if input, ok := MakeInputTakeAccountRecoveryCode(ctx, inputRawMessage); ok {
		state, err := s.Deps.ResetPassword.VerifyCode(ctx, input.AccountRecoveryCode)
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
		newFlowOutput, err := s.CreateNewFlow(ctx, flow, &SessionOptions{})
		if err != nil {
			return "", err
		}
		return newFlowOutput.Flow.StateToken, nil

	}
	return "", ErrFlowNotFound
}

func (s *Service) getSessionAndUpdateContext(ctx context.Context, flowID string) (context.Context, *Session, error) {
	session, err := s.Store.GetSession(ctx, flowID)
	if err != nil {
		return ctx, nil, err
	}

	ctx = session.MakeContext(ctx, s.Deps)

	return ctx, session, nil
}
