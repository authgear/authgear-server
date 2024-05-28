package webapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/base32"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

type AuthflowWechatCallbackData struct {
	State            string `json:"state"`
	Code             string `json:"code,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type AuthflowFinishedUIScreenData struct {
	FlowType          authflow.FlowType `json:"flow_type,omitempty"`
	FinishRedirectURI string            `json:"finish_redirect_uri,omitempty"`
}

type AuthflowDelayedUIScreenData struct {
	TargetResult *Result `json:"target_result,omitempty"`
}

const AuthflowQueryKey = "x_step"

// Authflow remembers all seen screens. The screens could come from more than 1 flow.
// We intentionally DO NOT clear screens when a different flow is created.
// As long as the browser has a reference to x_step, a screen can be retrieved.
// This design is important to ensure traversing browser history will not cause flow not found error.
// See https://github.com/authgear/authgear-server/issues/3452
type Authflow struct {
	// AllScreens is x_step => screen.
	AllScreens map[string]*AuthflowScreen `json:"all_screens,omitempty"`
}

func (s *Session) RememberScreen(screen *AuthflowScreen) {
	if s.Authflow == nil {
		s.Authflow = &Authflow{}
	}

	if s.Authflow.AllScreens == nil {
		s.Authflow.AllScreens = make(map[string]*AuthflowScreen)
	}

	s.Authflow.AllScreens[screen.StateToken.XStep] = screen
}

// AuthflowStateToken pairs x_step with its underlying state_token.
type AuthflowStateToken struct {
	XStep      string `json:"x_step"`
	StateToken string `json:"state_token"`
}

func NewAuthflowStateToken(flowResponse *authflow.FlowResponse) *AuthflowStateToken {
	xStep := newXStep()
	return &AuthflowStateToken{
		XStep:      xStep,
		StateToken: flowResponse.StateToken,
	}
}

func newXStep() string {
	const (
		idAlphabet string = base32.Alphabet
		idLength   int    = 32
	)
	return corerand.StringWithAlphabet(idLength, idAlphabet, corerand.SecureRand)
}

// AuthflowScreen represents a screen in the webapp.
// A screen typically corresponds to a step in an authflow.
// Some steps in an authflow can have branches.
// In order to be able to switch between branches, we need to remember the state that has branches.
type AuthflowScreen struct {
	// Store FinishedUIScreenData when the flow is finish
	FinishedUIScreenData *AuthflowFinishedUIScreenData `json:"finished_ui_screen_data,omitempty"`
	// Store DelayedUIScreenData when injecting screen between two steps
	DelayedUIScreenData *AuthflowDelayedUIScreenData `json:"delayed_ui_screen_data,omitempty"`
	// PreviousXStep is the x_step of the screen that leads to this screen.
	PreviousXStep string `json:"previous_x_step,omitempty"`
	// PreviousInput is the input that leads to this screen.
	// It can be nil.
	PreviousInput map[string]interface{} `json:"previous_input,omitempty"`
	// StateToken is always present.
	StateToken *AuthflowStateToken `json:"state_token,omitempty"`
	// BranchStateToken is only present when the underlying authflow step has branches.
	BranchStateToken *AuthflowStateToken `json:"branch_state_token,omitempty"`
	// TakenBranchIndex tracks the taken branch.
	TakenBranchIndex *int `json:"taken_branch_index,omitempty"`
	// TakenChannel tracks the taken channel.
	TakenChannel model.AuthenticatorOOBChannel `json:"taken_channel,omitempty"`
	// WechatCallbackData is only relevant for wechat login.
	WechatCallbackData *AuthflowWechatCallbackData `json:"wechat_callback_data,omitempty"`
}

func newAuthflowScreen(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	switch {
	case flowResponse.Action.Type == authflow.FlowActionTypeFinished:
		bytes, err := json.Marshal(flowResponse.Action.Data)
		if err != nil {
			panic(err)
		}

		var data map[string]interface{}
		err = json.Unmarshal(bytes, &data)
		if err != nil {
			panic(err)
		}
		finishRedirectURI, _ := data["finish_redirect_uri"].(string)
		state := NewAuthflowStateToken(flowResponse)
		screen := &AuthflowScreen{
			FinishedUIScreenData: &AuthflowFinishedUIScreenData{
				FlowType:          flowResponse.Type,
				FinishRedirectURI: finishRedirectURI,
			},
			PreviousXStep: previousXStep,
			PreviousInput: previousInput,
			StateToken:    state,
		}
		return screen
	default:
		switch flowResponse.Type {
		case authflow.FlowTypeSignup:
			return newAuthflowScreenSignup(flowResponse, previousXStep, previousInput)
		case authflow.FlowTypePromote:
			return newAuthflowScreenPromote(flowResponse, previousXStep, previousInput)
		case authflow.FlowTypeLogin:
			return newAuthflowScreenLogin(flowResponse, previousXStep, previousInput)
		case authflow.FlowTypeSignupLogin:
			return newAuthflowScreenSignupLogin(flowResponse, previousXStep, previousInput)
		case authflow.FlowTypeReauth:
			return newAuthflowScreenReauth(flowResponse, previousXStep, previousInput)
		case authflow.FlowTypeAccountRecovery:
			return newAuthflowScreenAccountRecovery(flowResponse, previousXStep, previousInput)
		default:
			panic(fmt.Errorf("unexpected flow type: %v", flowResponse.Type))
		}
	}
}

func newAuthflowScreenSignup(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	return newAuthflowScreenSignupPromote(flowResponse, previousXStep, previousInput)
}

func newAuthflowScreenPromote(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	return newAuthflowScreenSignupPromote(flowResponse, previousXStep, previousInput)
}

func newAuthflowScreenSignupPromote(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	state := NewAuthflowStateToken(flowResponse)
	screen := &AuthflowScreen{
		PreviousXStep: previousXStep,
		PreviousInput: previousInput,
		StateToken:    state,
	}
	switch config.AuthenticationFlowStepType(flowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// identify contains branches.
		screen.BranchStateToken = state
	case config.AuthenticationFlowStepTypeCreateAuthenticator:
		switch flowResponse.Action.Data.(type) {
		case declarative.IntentSignupFlowStepCreateAuthenticatorData:
			// create_authenticator contains branches in this step
			screen.BranchStateToken = state
		}
	case config.AuthenticationFlowStepTypeVerify:
		// verify MAY contain branches.
		if _, ok := flowResponse.Action.Data.(declarative.SelectOOBOTPChannelsData); ok {
			screen.BranchStateToken = state
		}
	case config.AuthenticationFlowStepTypeFillInUserProfile:
		// fill_in_user_profile contains NO branches.
		break
	case config.AuthenticationFlowStepTypeViewRecoveryCode:
		// view_recovery_code contains NO branches.
		break
	case config.AuthenticationFlowStepTypePromptCreatePasskey:
		// prompt_create_passkey contains NO branches.
		break
	default:
		panic(fmt.Errorf("unexpected action type: %v", flowResponse.Action.Type))
	}

	return screen
}

func newAuthflowScreenLogin(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	state := NewAuthflowStateToken(flowResponse)
	screen := &AuthflowScreen{
		PreviousXStep: previousXStep,
		PreviousInput: previousInput,
		StateToken:    state,
	}

	switch config.AuthenticationFlowStepType(flowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// identify contains branches.
		screen.BranchStateToken = state
	case config.AuthenticationFlowStepTypeAuthenticate:
		// authenticate contains branches.
		screen.BranchStateToken = state
	case config.AuthenticationFlowStepTypeCheckAccountStatus:
		// check_account_status contains NO branches.
		break
	case config.AuthenticationFlowStepTypeTerminateOtherSessions:
		// terminate_other_sessions contains NO branches.
		break
	case config.AuthenticationFlowStepTypeChangePassword:
		// change_password contains NO branches.
		break
	case config.AuthenticationFlowStepTypePromptCreatePasskey:
		// prompt_create_passkey contains NO branches.
		break
	default:
		panic(fmt.Errorf("unexpected action type: %v", flowResponse.Action.Type))
	}

	return screen
}

func newAuthflowScreenReauth(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	state := NewAuthflowStateToken(flowResponse)
	screen := &AuthflowScreen{
		PreviousXStep: previousXStep,
		PreviousInput: previousInput,
		StateToken:    state,
	}

	switch config.AuthenticationFlowStepType(flowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// identify contains branches.
		screen.BranchStateToken = state
	case config.AuthenticationFlowStepTypeAuthenticate:
		// authenticate contains branches.
		screen.BranchStateToken = state
	default:
		panic(fmt.Errorf("unexpected action type: %v", flowResponse.Action.Type))
	}

	return screen
}

func newAuthflowScreenSignupLogin(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	state := NewAuthflowStateToken(flowResponse)
	screen := &AuthflowScreen{
		PreviousXStep: previousXStep,
		PreviousInput: previousInput,
		StateToken:    state,
	}

	switch config.AuthenticationFlowStepType(flowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// identify contains branches.
		screen.BranchStateToken = state
	default:
		panic(fmt.Errorf("unexpected action type: %v", flowResponse.Action.Type))
	}

	return screen
}

func newAuthflowScreenAccountRecovery(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	state := NewAuthflowStateToken(flowResponse)
	screen := &AuthflowScreen{
		PreviousXStep: previousXStep,
		PreviousInput: previousInput,
		StateToken:    state,
	}

	switch config.AuthenticationFlowStepType(flowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// identify contains branches, but it is not important.
		break
	case config.AuthenticationFlowStepTypeSelectDestination:
		// select_destination contains branches, but it is not important.
		break
	case config.AuthenticationFlowStepTypeVerifyAccountRecoveryCode:
		// verify_account_recovery_code contains NO branches.
		break
	case config.AuthenticationFlowStepTypeResetPassword:
		// reset_password contains NO branches.
		break
	default:
		panic(fmt.Errorf("unexpected action type: %v", flowResponse.Action.Type))
	}

	return screen
}

type AuthflowScreenWithFlowResponse struct {
	Screen                       *AuthflowScreen
	StateTokenFlowResponse       *authflow.FlowResponse
	BranchStateTokenFlowResponse *authflow.FlowResponse
}

func NewAuthflowScreenWithResult(previousXStep string, targetResult *Result) *AuthflowScreen {
	return &AuthflowScreen{
		DelayedUIScreenData: &AuthflowDelayedUIScreenData{
			TargetResult: targetResult,
		},
		PreviousXStep: previousXStep,
		StateToken: &AuthflowStateToken{
			XStep:      newXStep(),
			StateToken: "",
		},
	}
}

func NewAuthflowScreenWithFlowResponse(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreenWithFlowResponse {
	screen := newAuthflowScreen(flowResponse, previousXStep, previousInput)
	screenWithResponse := &AuthflowScreenWithFlowResponse{
		Screen:                 screen,
		StateTokenFlowResponse: flowResponse,
	}
	if screen.BranchStateToken != nil {
		screenWithResponse.BranchStateTokenFlowResponse = flowResponse
	}
	return screenWithResponse
}

func UpdateAuthflowScreenWithFlowResponse(screen *AuthflowScreenWithFlowResponse, flowResponse *authflow.FlowResponse) *AuthflowScreenWithFlowResponse {
	clonedScreenWithFlowResponse := *screen
	clonedScreenWithFlowResponse.StateTokenFlowResponse = flowResponse

	clonedScreen := *clonedScreenWithFlowResponse.Screen
	clonedScreen.StateToken = &AuthflowStateToken{
		XStep:      screen.Screen.StateToken.XStep,
		StateToken: flowResponse.StateToken,
	}
	clonedScreenWithFlowResponse.Screen = &clonedScreen

	return &clonedScreenWithFlowResponse
}

func (s *AuthflowScreenWithFlowResponse) HasBranchToTake() bool {
	return s.Screen.BranchStateToken != nil && (s.Screen.TakenBranchIndex == nil && s.Screen.TakenChannel == "")
}

type TakeBranchResult interface {
	takeBranchResult()
}

type TakeBranchResultSimple struct {
	Screen *AuthflowScreenWithFlowResponse
}

func (TakeBranchResultSimple) takeBranchResult() {}

type TakeBranchResultInputRetryHandler func(err error) (nextInput interface{})

type TakeBranchResultInput struct {
	Input                 interface{}
	NewAuthflowScreenFull func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse
	OnRetry               *TakeBranchResultInputRetryHandler
}

func (TakeBranchResultInput) takeBranchResult() {}

type TakeBranchOptions struct {
	DisableFallbackToSMS bool
}

func (s *AuthflowScreenWithFlowResponse) InheritTakenBranchState(from *AuthflowScreenWithFlowResponse) {
	if from.BranchStateTokenFlowResponse != nil {
		s.BranchStateTokenFlowResponse = from.BranchStateTokenFlowResponse
		s.Screen.BranchStateToken = from.Screen.BranchStateToken
		s.Screen.TakenBranchIndex = from.Screen.TakenBranchIndex
		s.Screen.TakenChannel = from.Screen.TakenChannel
	}
}

func (s *AuthflowScreenWithFlowResponse) TakeBranch(index int, channel model.AuthenticatorOOBChannel, options *TakeBranchOptions) TakeBranchResult {
	switch s.StateTokenFlowResponse.Type {
	case authflow.FlowTypeSignup:
		return s.takeBranchSignup(index, channel, options)
	case authflow.FlowTypePromote:
		return s.takeBranchPromote(index, channel, options)
	case authflow.FlowTypeLogin:
		return s.takeBranchLogin(index, channel, options)
	case authflow.FlowTypeSignupLogin:
		return s.takeBranchSignupLogin(index, options)
	case authflow.FlowTypeReauth:
		return s.takeBranchReauth(index, channel, options)
	case authflow.FlowTypeAccountRecovery:
		return s.takeBranchAccountRecovery(index, options)
	default:
		panic(fmt.Errorf("unexpected flow type: %v", s.StateTokenFlowResponse.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchSignup(index int, channel model.AuthenticatorOOBChannel, options *TakeBranchOptions) TakeBranchResult {
	return s.takeBranchSignupPromote(index, channel, options)
}

func (s *AuthflowScreenWithFlowResponse) takeBranchPromote(index int, channel model.AuthenticatorOOBChannel, options *TakeBranchOptions) TakeBranchResult {
	return s.takeBranchSignupPromote(index, channel, options)
}

func (s *AuthflowScreenWithFlowResponse) takeBranchSignupPromote(index int, channel model.AuthenticatorOOBChannel, options *TakeBranchOptions) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index, channel)
	case config.AuthenticationFlowStepTypeCreateAuthenticator:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.IntentSignupFlowStepCreateAuthenticatorData)
		option := data.Options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			// Password branches can be taken by setting index.
			return s.takeBranchResultSimple(index, channel)
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			// This branch requires input to take.
			input := map[string]interface{}{
				"authentication": "secondary_totp",
			}
			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse {
					var emptyChannel model.AuthenticatorOOBChannel
					isContinuation := func(flowResponse *authflow.FlowResponse) bool {
						return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator) &&
							flowResponse.Action.Authentication == config.AuthenticationFlowAuthenticationSecondaryTOTP
					}

					return s.makeScreenForTakenBranch(flowResponse, input, &index, emptyChannel, isContinuation)
				},
			}
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			if channel == "" {
				channel = option.Channels[0]
			}
			inputFactory := func(c model.AuthenticatorOOBChannel) map[string]interface{} {
				return map[string]interface{}{
					"authentication": option.Authentication,
					"channel":        c,
				}
			}
			input := inputFactory(channel)
			onFailureHandler := s.makeFallbackToSMSFromWhatsappRetryHandler(
				inputFactory,
				option.Channels,
				options.DisableFallbackToSMS,
			)
			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflow.FlowResponse) bool {
						return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator) &&
							flowResponse.Action.Authentication == option.Authentication
					}
					takenChannel := channel
					if d, ok := flowResponse.Action.Data.(declarative.VerifyOOBOTPData); ok {
						takenChannel = d.Channel
					}

					screen := s.makeScreenForTakenBranch(flowResponse, input, &index, takenChannel, isContinuation)
					return screen
				},
				OnRetry: &onFailureHandler,
			}

		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			if channel == "" {
				channel = option.Channels[0]
			}
			return s.takeBranchResultSimple(index, channel)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case config.AuthenticationFlowStepTypeVerify:
		// If we ever reach here, this means we have to choose channels.
		data := s.StateTokenFlowResponse.Action.Data.(declarative.SelectOOBOTPChannelsData)
		if channel == "" {
			channel = data.Channels[0]
		}
		inputFactory := func(c model.AuthenticatorOOBChannel) map[string]interface{} {
			return map[string]interface{}{
				"channel": c,
			}
		}
		input := inputFactory(channel)
		onFailureHandler := s.makeFallbackToSMSFromWhatsappRetryHandler(
			inputFactory,
			data.Channels,
			options.DisableFallbackToSMS,
		)
		return TakeBranchResultInput{
			Input: input,
			NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse {
				var nilIndex *int
				isContinuation := func(flowResponse *authflow.FlowResponse) bool {
					return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowSignupFlowStepTypeVerify)
				}
				takenChannel := channel
				if d, ok := flowResponse.Action.Data.(declarative.VerifyOOBOTPData); ok {
					takenChannel = d.Channel
				}
				screen := s.makeScreenForTakenBranch(flowResponse, input, nilIndex, takenChannel, isContinuation)
				return screen
			},
			OnRetry: &onFailureHandler,
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchLogin(index int, channel model.AuthenticatorOOBChannel, options *TakeBranchOptions) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index, channel)
	case config.AuthenticationFlowStepTypeAuthenticate:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.StepAuthenticateData)
		option := data.Options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			fallthrough
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			// All these can take the branch simply by setting index.
			return s.takeBranchResultSimple(index, channel)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			// This branch requires input to take.
			if channel == "" {
				channel = option.Channels[0]
			}

			inputFactory := func(c model.AuthenticatorOOBChannel) map[string]interface{} {
				return map[string]interface{}{
					"authentication": option.Authentication,
					"index":          index,
					"channel":        c,
				}
			}
			input := inputFactory(channel)
			onFailureHandler := s.makeFallbackToSMSFromWhatsappRetryHandler(
				inputFactory,
				option.Channels,
				options.DisableFallbackToSMS,
			)

			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflow.FlowResponse) bool {
						return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowLoginFlowStepTypeAuthenticate) && flowResponse.Action.Authentication == option.Authentication
					}
					takenChannel := channel
					switch d := flowResponse.Action.Data.(type) {
					case declarative.VerifyOOBOTPData:
						takenChannel = d.Channel
					default:
						// Skip if we do not have the channel.
					}

					screen := s.makeScreenForTakenBranch(flowResponse, input, &index, takenChannel, isContinuation)
					return screen
				},
				OnRetry: &onFailureHandler,
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchReauth(index int, channel model.AuthenticatorOOBChannel, options *TakeBranchOptions) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, id_token is used.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index, channel)
	case config.AuthenticationFlowStepTypeAuthenticate:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.StepAuthenticateData)
		option := data.Options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			// All these can take the branch simply by setting index.
			return s.takeBranchResultSimple(index, channel)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			// This branch requires input to take.
			if channel == "" {
				channel = option.Channels[0]
			}

			inputFactory := func(c model.AuthenticatorOOBChannel) map[string]interface{} {
				return map[string]interface{}{
					"authentication": option.Authentication,
					"index":          index,
					"channel":        c,
				}
			}
			input := inputFactory(channel)
			onFailureHandler := s.makeFallbackToSMSFromWhatsappRetryHandler(
				inputFactory,
				option.Channels,
				options.DisableFallbackToSMS,
			)

			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflow.FlowResponse) bool {
						return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowStepTypeAuthenticate) && flowResponse.Action.Authentication == option.Authentication
					}
					takenChannel := channel
					switch d := flowResponse.Action.Data.(type) {
					case declarative.VerifyOOBOTPData:
						takenChannel = d.Channel
					default:
						// Skip if we do not have the channel.
					}

					screen := s.makeScreenForTakenBranch(flowResponse, input, &index, takenChannel, isContinuation)
					return screen
				},
				OnRetry: &onFailureHandler,
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchSignupLogin(index int, options *TakeBranchOptions) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index, "")
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchAccountRecovery(index int, options *TakeBranchOptions) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index, "")
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchResultSimple(index int, channel model.AuthenticatorOOBChannel) TakeBranchResultSimple {
	xStep := s.Screen.StateToken.XStep
	screen := NewAuthflowScreenWithFlowResponse(s.StateTokenFlowResponse, xStep, nil)
	screen.Screen.BranchStateToken = s.Screen.StateToken
	screen.BranchStateTokenFlowResponse = s.BranchStateTokenFlowResponse
	screen.Screen.TakenBranchIndex = &index
	screen.Screen.TakenChannel = channel
	return TakeBranchResultSimple{
		Screen: screen,
	}
}

func (s *AuthflowScreenWithFlowResponse) makeScreenForTakenBranch(
	flowResponse *authflow.FlowResponse,
	input map[string]interface{},
	index *int,
	channel model.AuthenticatorOOBChannel,
	isContinuation func(flowResponse *authflow.FlowResponse) bool,
) *AuthflowScreenWithFlowResponse {
	// Sometimes, when we take a branch, the branch ends immediately.
	// In that case, we consider the screen as another branching point.
	// One particular case is the following signup flow
	// 1. identify with email
	// 2. email is required to verify
	// 3. primary_oob_otp_email is taken. But verification was done in Step 2, so this step ends immediately.
	// 4. secondary_totp is pending to be taken. This is another branching point.
	//
	// Therefore, we need to tell if the screen created from flowResponse is
	// the continuation of s.
	if isContinuation(flowResponse) {
		xStep := s.Screen.StateToken.XStep
		screen := NewAuthflowScreenWithFlowResponse(flowResponse, xStep, input)
		screen.Screen.BranchStateToken = s.Screen.StateToken
		screen.BranchStateTokenFlowResponse = s.StateTokenFlowResponse
		screen.Screen.TakenBranchIndex = index
		screen.Screen.TakenChannel = channel
		return screen
	} else {
		xStep := s.Screen.StateToken.XStep
		screen := NewAuthflowScreenWithFlowResponse(flowResponse, xStep, input)
		return screen
	}
}

type Navigator interface {
	Navigate(screen *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result)
}

func (s *AuthflowScreenWithFlowResponse) Navigate(navigator Navigator, r *http.Request, webSessionID string, result *Result) {
	navigator.Navigate(s, r, webSessionID, result)
}

func (s *AuthflowScreenWithFlowResponse) Advance(p string, result *Result) {
	q := url.Values{}
	q.Set(AuthflowQueryKey, s.Screen.StateToken.XStep)
	u, _ := url.Parse(p)
	u.RawQuery = q.Encode()

	result.NavigationAction = "advance"
	result.RedirectURI = u.String()
}

func (s *AuthflowScreenWithFlowResponse) makeFallbackToSMSFromWhatsappRetryHandler(
	inputFactory func(channel model.AuthenticatorOOBChannel) map[string]interface{},
	channels []model.AuthenticatorOOBChannel,
	disableFallbackToSMS bool) TakeBranchResultInputRetryHandler {
	return func(err error) interface{} {
		if disableFallbackToSMS {
			return nil
		}
		if !errors.Is(err, otp.ErrInvalidWhatsappUser) {
			return nil
		}
		smsChannelIdx := -1
		for idx, c := range channels {
			if c == model.AuthenticatorOOBChannelSMS {
				smsChannelIdx = idx
			}
		}
		if smsChannelIdx == -1 {
			return nil
		}
		return inputFactory(channels[smsChannelIdx])
	}
}
