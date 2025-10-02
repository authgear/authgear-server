package webapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflowv2viewmodels "github.com/authgear/authgear-server/pkg/auth/handler/webapp/authflowv2/viewmodels"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/whatsapp"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
	"github.com/authgear/authgear-server/pkg/util/base32"
	"github.com/authgear/authgear-server/pkg/util/clock"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

type AuthflowWechatCallbackData struct {
	State            string                        `json:"state"`
	WebappOAuthState *webappoauth.WebappOAuthState `json:"webapp_oauth_state"`
	Query            string                        `json:"query"`
}

type AuthflowFinishedUIScreenData struct {
	FlowType          authflow.FlowType `json:"flow_type,omitempty"`
	FinishRedirectURI string            `json:"finish_redirect_uri,omitempty"`
}

type AuthflowDelayedUIScreenData struct {
	TargetResult *Result `json:"target_result,omitempty"`
}

const AuthflowQueryKey = "x_step"

const messageDeliveryStatusMaxWaitTime = 30 * time.Second

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
	// IsBotProtectionRequired will be used to determine whether to navigate to bot protection verification screen.
	IsBotProtectionRequired bool `json:"is_bot_protection_required,omitempty"`

	// In some cases, we intentionally add screens between steps, so the path may not match
	SkipPathCheck bool `json:"skip_path_check,omitempty"`

	// viewmodels used in specific screens
	OAuthProviderDemoCredentialViewModel *authflowv2viewmodels.OAuthProviderDemoCredentialViewModel `json:"oauth_provider_demo_credential,omitempty"`
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
		case declarative.CreateAuthenticatorData:
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

func NewAuthflowDelayedScreenWithResult(
	sourceScreen *AuthflowScreen, targetResult *Result) *AuthflowScreen {
	return &AuthflowScreen{
		DelayedUIScreenData: &AuthflowDelayedUIScreenData{
			TargetResult: targetResult,
		},
		PreviousXStep: sourceScreen.StateToken.XStep,
		StateToken: &AuthflowStateToken{
			XStep:      newXStep(),
			StateToken: sourceScreen.StateToken.StateToken,
		},
		// Delayed screen does not have a corresponding authflow step, skip the path check
		SkipPathCheck: true,
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
type TakeBranchOutputTransformer func(ctx context.Context, output *authflow.ServiceOutput, err error, deps TransformerDependencies) (*authflow.ServiceOutput, error)

type AuthflowService interface {
	Get(ctx context.Context, stateToken string) (*authflow.ServiceOutput, error)
}

type TransformerDependencies struct {
	Clock     clock.Clock
	Authflows AuthflowService
}
type TakeBranchResultInput struct {
	Input                 map[string]interface{}
	NewAuthflowScreenFull func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse
	TransformOutput       *TakeBranchOutputTransformer
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

type TakeBranchInput struct {
	Index   int
	Channel model.AuthenticatorOOBChannel

	// bot protection specific inputs
	BotProtectionProviderType     string
	BotProtectionProviderResponse string
}

func (i *TakeBranchInput) HasBotProtectionInput() bool {
	return i != nil && i.BotProtectionProviderType != "" && i.BotProtectionProviderResponse != ""
}

func (s *AuthflowScreenWithFlowResponse) TakeBranch(input *TakeBranchInput, options *TakeBranchOptions) TakeBranchResult {
	switch s.StateTokenFlowResponse.Type {
	case authflow.FlowTypeSignup:
		return s.takeBranchSignup(input, options)
	case authflow.FlowTypePromote:
		return s.takeBranchPromote(input, options)
	case authflow.FlowTypeLogin:
		return s.takeBranchLogin(input, options)
	case authflow.FlowTypeSignupLogin:
		return s.takeBranchSignupLogin(input.Index, options)
	case authflow.FlowTypeReauth:
		return s.takeBranchReauth(input, options)
	case authflow.FlowTypeAccountRecovery:
		return s.takeBranchAccountRecovery(input.Index, options)
	default:
		panic(fmt.Errorf("unexpected flow type: %v", s.StateTokenFlowResponse.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchSignup(input *TakeBranchInput, options *TakeBranchOptions) TakeBranchResult {
	return s.takeBranchSignupPromote(input, options)
}

func (s *AuthflowScreenWithFlowResponse) takeBranchPromote(input *TakeBranchInput, options *TakeBranchOptions) TakeBranchResult {
	return s.takeBranchSignupPromote(input, options)
}

func (s *AuthflowScreenWithFlowResponse) takeBranchSignupPromote(input *TakeBranchInput, options *TakeBranchOptions) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(input, false)
	case config.AuthenticationFlowStepTypeCreateAuthenticator:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.CreateAuthenticatorData)
		return s.takeBranchCreateAuthenticator(input, config.AuthenticationFlowStepTypeCreateAuthenticator, options, data.Options[input.Index])
	case config.AuthenticationFlowStepTypeVerify:
		// If we ever reach here, this means we have to choose channels.
		data := s.StateTokenFlowResponse.Action.Data.(declarative.SelectOOBOTPChannelsData)
		if input.Channel == "" {
			input.Channel = data.Channels[0]
		}
		inputFactory := func(c model.AuthenticatorOOBChannel) map[string]interface{} {
			return map[string]interface{}{
				"channel": c,
			}
		}
		resultInput := inputFactory(input.Channel)
		inputOptions := s.makeVerifyOOBOTPInputOptions(
			inputFactory,
			data.Channels,
			options.DisableFallbackToSMS,
		)
		return TakeBranchResultInput{
			Input: resultInput,
			NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse {
				var nilIndex *int
				isContinuation := func(flowResponse *authflow.FlowResponse) bool {
					return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowSignupFlowStepTypeVerify)
				}
				takenChannel := input.Channel
				if d, ok := flowResponse.Action.Data.(declarative.VerifyOOBOTPData); ok {
					takenChannel = d.Channel
				}
				screen := s.makeScreenForTakenBranch(flowResponse, resultInput, nilIndex, takenChannel, isContinuation)
				return screen
			},
			OnRetry:         inputOptions.RetryHandler,
			TransformOutput: inputOptions.OutputTransformer,
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchLoginAuthenticate(input *TakeBranchInput, options *TakeBranchOptions) TakeBranchResult {
	switch data := s.StateTokenFlowResponse.Action.Data.(type) {
	case declarative.StepAuthenticateData:
		option := data.Options[input.Index]
		switch option.Authentication {
		case model.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryPassword:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryTOTP:
			fallthrough
		case model.AuthenticationFlowAuthenticationRecoveryCode:
			fallthrough
		case model.AuthenticationFlowAuthenticationPrimaryPasskey:
			// All these can take the branch simply by setting index.
			return s.takeBranchResultSimple(input, false)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			// This branch requires input to take.
			if input.Channel == "" {
				input.Channel = option.Channels[0]
			}
			// Below clause takes place only when index=0 branch is OOBOTP, and bot protection not in input
			// otherwise, the bot protection is fed to auto-taken OOBOTP branch too
			if option.BotProtection.IsRequired() && !input.HasBotProtectionInput() {
				return s.takeBranchResultSimple(input, true)
			}

			inputFactory := func(c model.AuthenticatorOOBChannel) map[string]interface{} {
				out := map[string]interface{}{
					"authentication": option.Authentication,
					"index":          input.Index,
					"channel":        c,
				}
				if input.HasBotProtectionInput() {
					bp := map[string]interface{}{
						"type":     input.BotProtectionProviderType,
						"response": input.BotProtectionProviderResponse,
					}
					out["bot_protection"] = bp
				}
				return out
			}
			resultInput := inputFactory(input.Channel)
			inputOptions := s.makeVerifyOOBOTPInputOptions(
				inputFactory,
				option.Channels,
				options.DisableFallbackToSMS,
			)

			return TakeBranchResultInput{
				Input: resultInput,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflow.FlowResponse) bool {
						return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowLoginFlowStepTypeAuthenticate) && flowResponse.Action.Authentication == option.Authentication
					}
					takenChannel := input.Channel
					switch d := flowResponse.Action.Data.(type) {
					case declarative.VerifyOOBOTPData:
						takenChannel = d.Channel
					default:
						// Skip if we do not have the channel.
					}

					screen := s.makeScreenForTakenBranch(flowResponse, resultInput, &input.Index, takenChannel, isContinuation)
					return screen
				},
				OnRetry:         inputOptions.RetryHandler,
				TransformOutput: inputOptions.OutputTransformer,
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case declarative.VerifyOOBOTPData:
		channel := data.Channel
		return s.takeBranchResultSimple(&TakeBranchInput{Index: input.Index, Channel: channel}, false)
	default:
		panic(fmt.Errorf("unexpected data type: %T", s.StateTokenFlowResponse.Action.Data))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchLogin(input *TakeBranchInput, options *TakeBranchOptions) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(input, false)
	case config.AuthenticationFlowStepTypeAuthenticate:
		switch s.StateTokenFlowResponse.Action.Data.(type) {
		case declarative.CreateAuthenticatorData:
			data := s.StateTokenFlowResponse.Action.Data.(declarative.CreateAuthenticatorData)
			return s.takeBranchCreateAuthenticator(input, config.AuthenticationFlowStepTypeAuthenticate, options, data.Options[input.Index])
		case declarative.IntentCreateAuthenticatorTOTPData:
			return s.takeBranchResultSimple(input, false)
		case declarative.StepAuthenticateData:
			return s.takeBranchLoginAuthenticate(input, options)
		default:
			panic(fmt.Errorf("unexpected action data: %T", s.StateTokenFlowResponse.Action.Data))
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchReauth(input *TakeBranchInput, options *TakeBranchOptions) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, id_token is used.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(input, false)
	case config.AuthenticationFlowStepTypeAuthenticate:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.StepAuthenticateData)
		option := data.Options[input.Index]
		switch option.Authentication {
		case model.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryPassword:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryTOTP:
			fallthrough
		case model.AuthenticationFlowAuthenticationPrimaryPasskey:
			// All these can take the branch simply by setting index.
			return s.takeBranchResultSimple(input, false)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			// This branch requires input to take.
			if input.Channel == "" {
				input.Channel = option.Channels[0]
			}

			if option.BotProtection.IsRequired() {
				return s.takeBranchResultSimple(input, true)
			}

			inputFactory := func(c model.AuthenticatorOOBChannel) map[string]interface{} {
				return map[string]interface{}{
					"authentication": option.Authentication,
					"index":          input.Index,
					"channel":        c,
				}
			}
			resultInput := inputFactory(input.Channel)
			inputOptions := s.makeVerifyOOBOTPInputOptions(
				inputFactory,
				option.Channels,
				options.DisableFallbackToSMS,
			)

			return TakeBranchResultInput{
				Input: resultInput,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflow.FlowResponse) bool {
						return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowStepTypeAuthenticate) && flowResponse.Action.Authentication == option.Authentication
					}
					takenChannel := input.Channel
					switch d := flowResponse.Action.Data.(type) {
					case declarative.VerifyOOBOTPData:
						takenChannel = d.Channel
					default:
						// Skip if we do not have the channel.
					}

					screen := s.makeScreenForTakenBranch(flowResponse, resultInput, &input.Index, takenChannel, isContinuation)
					return screen
				},
				OnRetry:         inputOptions.RetryHandler,
				TransformOutput: inputOptions.OutputTransformer,
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
		return s.takeBranchResultSimple(&TakeBranchInput{Index: index, Channel: ""}, false)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchAccountRecovery(index int, options *TakeBranchOptions) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(&TakeBranchInput{Index: index, Channel: ""}, false)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchResultSimple(input *TakeBranchInput, botProtectionRequired bool) TakeBranchResultSimple {
	xStep := s.Screen.StateToken.XStep
	screen := NewAuthflowScreenWithFlowResponse(s.StateTokenFlowResponse, xStep, nil)
	screen.Screen.BranchStateToken = s.Screen.StateToken
	screen.BranchStateTokenFlowResponse = s.BranchStateTokenFlowResponse
	screen.Screen.TakenBranchIndex = &input.Index
	screen.Screen.TakenChannel = input.Channel
	screen.Screen.IsBotProtectionRequired = botProtectionRequired
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
	Navigate(ctx context.Context, screen *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result)
}

func (s *AuthflowScreenWithFlowResponse) Navigate(ctx context.Context, navigator Navigator, r *http.Request, webSessionID string, result *Result) {
	navigator.Navigate(ctx, s, r, webSessionID, result)
}

// Advance is for advancing to another page to drive the authflow.
func (s *AuthflowScreenWithFlowResponse) Advance(route string, result *Result) {
	s.advanceOrRedirect(route, result)
	result.NavigationAction = NavigationActionAdvance
}

// RedirectToFinish is a fix for https://linear.app/authgear/issue/DEV-1793/investigate-sign-in-directly-with-httpsaccountsportalauthgearcom-crash
// We need Turbo to visit /finish with a full browser redirect,
// so CSP and connect-src will not kick in.
func (s *AuthflowScreenWithFlowResponse) RedirectToFinish(route string, result *Result) {
	s.advanceOrRedirect(route, result)
	result.NavigationAction = NavigationActionRedirect
}

// advanceOrRedirect is for internal use. You use Advance or RedirectToFinish instead.
func (s *AuthflowScreenWithFlowResponse) advanceOrRedirect(route string, result *Result) {
	q := url.Values{}
	q.Set(AuthflowQueryKey, s.Screen.StateToken.XStep)
	u, _ := url.Parse(route)
	u.RawQuery = q.Encode()
	result.RedirectURI = u.String()
}

func (s *AuthflowScreenWithFlowResponse) AdvanceWithQuery(route string, result *Result, query url.Values) {
	q := url.Values{}
	q.Set(AuthflowQueryKey, s.Screen.StateToken.XStep)

	for k, v := range query {
		q[k] = v
	}

	u, _ := url.Parse(route)
	u.RawQuery = q.Encode()

	result.NavigationAction = NavigationActionAdvance
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
		if !apierrors.IsKind(err, whatsapp.InvalidWhatsappUser) &&
			!apierrors.IsKind(err, whatsapp.WhatsappUndeliverable) {
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

func (s *AuthflowScreenWithFlowResponse) makeVerifyOOBOTPOutputTransformer(channels []model.AuthenticatorOOBChannel) TakeBranchOutputTransformer {
	hasNonWhatsappChannels := false
	for _, c := range channels {
		if c == model.AuthenticatorOOBChannelWhatsapp {
			continue
		}
		hasNonWhatsappChannels = true
	}
	return func(ctx context.Context, output *authflow.ServiceOutput, err error, deps TransformerDependencies) (*authflow.ServiceOutput, error) {
		if err != nil {
			// If there is error, make no changes to the output and error
			return output, err
		}
		if _, ok := output.FlowAction.Data.(declarative.VerifyOOBOTPData); !ok {
			// If not VerifyOOBOTPData, make no changes to the output and error
			return output, err
		} else if !hasNonWhatsappChannels {
			// No fallback available, show the otp screen no matter success or not
			return output, err
		} else {
			startTime := deps.Clock.NowUTC()

			for {
				now := deps.Clock.NowUTC()
				// Normally, the timeout should be handled by the whatsapp callback timeout.
				// Therefore we should always get failed / sent status within the configured timeout.
				// So this timeout is just as a last resort to break the loop to avoid waiting forever.
				if now.Sub(startTime) > messageDeliveryStatusMaxWaitTime {
					return nil, fmt.Errorf("timed out waiting for delivery_status")
				}

				data := output.FlowAction.Data.(declarative.VerifyOOBOTPData)
				switch data.DeliveryStatus {
				case model.OTPDeliveryStatusFailed:
					return output, data.DeliveryError
				case model.OTPDeliveryStatusSent:
					return output, err
				case model.OTPDeliveryStatusSending:
					{
						// Wait until sent or failed
						time.Sleep(500 * time.Millisecond)
						newOutput, err := deps.Authflows.Get(ctx, output.Flow.StateToken)
						if err != nil {
							return newOutput, err
						}
						_, ok := newOutput.FlowAction.Data.(declarative.VerifyOOBOTPData)
						if ok {
							output = newOutput
							continue
						} else {
							// The flow has advanced to the next step.
							return newOutput, nil
						}
					}
				}
			}
		}
	}
}

func (s *AuthflowScreenWithFlowResponse) makeVerifyOOBOTPInputOptions(
	inputFactory func(channel model.AuthenticatorOOBChannel) map[string]interface{},
	channels []model.AuthenticatorOOBChannel,
	disableFallbackToSMS bool,
) struct {
	OutputTransformer *TakeBranchOutputTransformer
	RetryHandler      *TakeBranchResultInputRetryHandler
} {
	var transformer TakeBranchOutputTransformer = s.makeVerifyOOBOTPOutputTransformer(channels)
	var retryHandler TakeBranchResultInputRetryHandler = s.makeFallbackToSMSFromWhatsappRetryHandler(
		inputFactory, channels, disableFallbackToSMS,
	)
	return struct {
		OutputTransformer *TakeBranchOutputTransformer
		RetryHandler      *TakeBranchResultInputRetryHandler
	}{
		OutputTransformer: &transformer,
		RetryHandler:      &retryHandler,
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchCreateAuthenticator(
	input *TakeBranchInput,
	expectedActionType config.AuthenticationFlowStepType,
	takeBranchOptions *TakeBranchOptions,
	selectedOption declarative.CreateAuthenticatorOptionForOutput,
) TakeBranchResult {
	switch selectedOption.Authentication {
	case model.AuthenticationFlowAuthenticationPrimaryPassword:
		fallthrough
	case model.AuthenticationFlowAuthenticationSecondaryPassword:
		// Password branches can be taken by setting index.
		return s.takeBranchResultSimple(input, false)
	case model.AuthenticationFlowAuthenticationSecondaryTOTP:
		// This branch requires input to take.
		resultInput := map[string]interface{}{
			"authentication": "secondary_totp",
		}
		return TakeBranchResultInput{
			Input: resultInput,
			NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse {
				var emptyChannel model.AuthenticatorOOBChannel
				isContinuation := func(flowResponse *authflow.FlowResponse) bool {
					return flowResponse.Action.Type == authflow.FlowActionType(expectedActionType) &&
						flowResponse.Action.Authentication == model.AuthenticationFlowAuthenticationSecondaryTOTP
				}

				return s.makeScreenForTakenBranch(flowResponse, resultInput, &input.Index, emptyChannel, isContinuation)
			},
		}
	case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
		fallthrough
	case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
		if input.Channel == "" {
			input.Channel = selectedOption.Channels[0]
		}
		inputFactory := func(c model.AuthenticatorOOBChannel) map[string]interface{} {
			return map[string]interface{}{
				"authentication": selectedOption.Authentication,
				"channel":        c,
			}
		}
		resultInput := inputFactory(input.Channel)
		inputOptions := s.makeVerifyOOBOTPInputOptions(
			inputFactory,
			selectedOption.Channels,
			takeBranchOptions.DisableFallbackToSMS,
		)
		return TakeBranchResultInput{
			Input: resultInput,
			NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse, retriedForError error) *AuthflowScreenWithFlowResponse {
				isContinuation := func(flowResponse *authflow.FlowResponse) bool {
					return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator) &&
						flowResponse.Action.Authentication == selectedOption.Authentication
				}
				takenChannel := input.Channel
				if d, ok := flowResponse.Action.Data.(declarative.VerifyOOBOTPData); ok {
					takenChannel = d.Channel
				}

				screen := s.makeScreenForTakenBranch(flowResponse, resultInput, &input.Index, takenChannel, isContinuation)
				return screen
			},
			OnRetry:         inputOptions.RetryHandler,
			TransformOutput: inputOptions.OutputTransformer,
		}

	case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
		fallthrough
	case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
		if input.Channel == "" {
			input.Channel = selectedOption.Channels[0]
		}
		return s.takeBranchResultSimple(&TakeBranchInput{
			Index:   input.Index,
			Channel: input.Channel,
		}, false)
	default:
		panic(fmt.Errorf("unexpected authentication: %v", selectedOption.Authentication))
	}
}
