package webapp

import (
	"encoding/base64"
	"encoding/json"
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

type AuthflowOAuthState struct {
	WebSessionID     string `json:"web_session_id"`
	XStep            string `json:"x_step"`
	ErrorRedirectURI string `json:"error_redirect_uri"`
}

func (s AuthflowOAuthState) Encode() string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

func DecodeAuthflowOAuthState(stateStr string) (*AuthflowOAuthState, error) {
	b, err := base64.RawURLEncoding.DecodeString(stateStr)
	if err != nil {
		return nil, err
	}

	var state AuthflowOAuthState
	err = json.Unmarshal(b, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

const AuthflowQueryKey = "x_step"

// Authflow stores the necessary information for webapp to run an authflow, including
// navigation, and branching.
type Authflow struct {
	// AllScreens is x_step => screen.
	AllScreens map[string]*AuthflowScreen `json:"all_screens,omitempty"`
}

func NewAuthflow(initialScreen *AuthflowScreenWithFlowResponse) *Authflow {
	af := &Authflow{
		AllScreens: map[string]*AuthflowScreen{},
	}
	af.RememberScreen(initialScreen)
	return af
}

func (f *Authflow) RememberScreen(screen *AuthflowScreenWithFlowResponse) {
	f.AllScreens[screen.Screen.StateToken.XStep] = screen.Screen
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
		// authentication contains branches.
		screen.BranchStateToken = state
	case config.AuthenticationFlowStepTypeVerify:
		// verify MAY contain branches.
		if _, ok := flowResponse.Action.Data.(declarative.IntentVerifyClaimData); ok {
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
		// identify contains branches.
		screen.BranchStateToken = state
	case config.AuthenticationFlowStepTypeSelectDestination:
		// select_destination contains branches.
		screen.BranchStateToken = state
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

type TakeBranchResultInput struct {
	Input                 interface{}
	NewAuthflowScreenFull func(flowResponse *authflow.FlowResponse) *AuthflowScreenWithFlowResponse
}

func (TakeBranchResultInput) takeBranchResult() {}

func (s *AuthflowScreenWithFlowResponse) TakeBranch(index int, channel model.AuthenticatorOOBChannel) TakeBranchResult {
	switch s.StateTokenFlowResponse.Type {
	case authflow.FlowTypeSignup:
		return s.takeBranchSignup(index, channel)
	case authflow.FlowTypePromote:
		return s.takeBranchPromote(index, channel)
	case authflow.FlowTypeLogin:
		return s.takeBranchLogin(index, channel)
	case authflow.FlowTypeSignupLogin:
		return s.takeBranchSignupLogin(index)
	case authflow.FlowTypeReauth:
		return s.takeBranchReauth(index, channel)
	case authflow.FlowTypeAccountRecovery:
		return s.takeBranchAccountRecovery(index)
	default:
		panic(fmt.Errorf("unexpected flow type: %v", s.StateTokenFlowResponse.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchSignup(index int, channel model.AuthenticatorOOBChannel) TakeBranchResult {
	return s.takeBranchSignupPromote(index, channel)
}

func (s *AuthflowScreenWithFlowResponse) takeBranchPromote(index int, channel model.AuthenticatorOOBChannel) TakeBranchResult {
	return s.takeBranchSignupPromote(index, channel)
}

func (s *AuthflowScreenWithFlowResponse) takeBranchSignupPromote(index int, channel model.AuthenticatorOOBChannel) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index)
	case config.AuthenticationFlowStepTypeCreateAuthenticator:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.IntentSignupFlowStepCreateAuthenticatorData)
		option := data.Options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			// Password branches can be taken by setting index.
			return s.takeBranchResultSimple(index)
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			// This branch requires input to take.
			input := map[string]interface{}{
				"authentication": "secondary_totp",
			}
			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse) *AuthflowScreenWithFlowResponse {
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
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			if channel == "" {
				channel = option.Channels[0]
			}
			input := map[string]interface{}{
				"authentication": option.Authentication,
				"channel":        channel,
			}
			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflow.FlowResponse) bool {
						return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowSignupFlowStepTypeCreateAuthenticator) &&
							flowResponse.Action.Authentication == option.Authentication
					}

					return s.makeScreenForTakenBranch(flowResponse, input, &index, channel, isContinuation)
				},
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case config.AuthenticationFlowStepTypeVerify:
		// If we ever reach here, this means we have to choose channels.
		data := s.StateTokenFlowResponse.Action.Data.(declarative.IntentVerifyClaimData)
		if channel == "" {
			channel = data.Channels[0]
		}
		input := map[string]interface{}{
			"channel": channel,
		}
		return TakeBranchResultInput{
			Input: input,
			NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse) *AuthflowScreenWithFlowResponse {
				var nilIndex *int
				isContinuation := func(flowResponse *authflow.FlowResponse) bool {
					return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowSignupFlowStepTypeVerify)
				}
				return s.makeScreenForTakenBranch(flowResponse, input, nilIndex, channel, isContinuation)
			},
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchLogin(index int, channel model.AuthenticatorOOBChannel) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index)
	case config.AuthenticationFlowStepTypeAuthenticate:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.IntentLoginFlowStepAuthenticateData)
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
			return s.takeBranchResultSimple(index)
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
			input := map[string]interface{}{
				"authentication": option.Authentication,
				"index":          index,
				"channel":        channel,
			}

			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflow.FlowResponse) bool {
						return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowLoginFlowStepTypeAuthenticate) && flowResponse.Action.Authentication == option.Authentication
					}

					return s.makeScreenForTakenBranch(flowResponse, input, &index, channel, isContinuation)

				},
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchReauth(index int, channel model.AuthenticatorOOBChannel) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, id_token is used.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index)
	case config.AuthenticationFlowStepTypeAuthenticate:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.IntentReauthFlowStepAuthenticateData)
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
			return s.takeBranchResultSimple(index)
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
			input := map[string]interface{}{
				"authentication": option.Authentication,
				"index":          index,
				"channel":        channel,
			}

			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflow.FlowResponse) bool {
						return flowResponse.Action.Type == authflow.FlowActionType(config.AuthenticationFlowStepTypeAuthenticate) && flowResponse.Action.Authentication == option.Authentication
					}

					return s.makeScreenForTakenBranch(flowResponse, input, &index, channel, isContinuation)

				},
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchSignupLogin(index int) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchAccountRecovery(index int) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchResultSimple(index int) TakeBranchResultSimple {
	xStep := s.Screen.StateToken.XStep
	screen := NewAuthflowScreenWithFlowResponse(s.StateTokenFlowResponse, xStep, nil)
	screen.Screen.BranchStateToken = s.Screen.StateToken
	screen.BranchStateTokenFlowResponse = s.BranchStateTokenFlowResponse
	screen.Screen.TakenBranchIndex = &index
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

func (s *AuthflowScreenWithFlowResponse) Navigate(r *http.Request, webSessionID string, result *Result) {
	if s.HasBranchToTake() {
		panic(fmt.Errorf("expected screen to have its branches taken"))
	}

	switch s.StateTokenFlowResponse.Type {
	case authflow.FlowTypeSignup:
		s.navigateSignup(r, webSessionID, result)
	case authflow.FlowTypePromote:
		s.navigatePromote(r, webSessionID, result)
	case authflow.FlowTypeLogin:
		s.navigateLogin(r, webSessionID, result)
	case authflow.FlowTypeSignupLogin:
		s.navigateSignupLogin(r, webSessionID, result)
	case authflow.FlowTypeReauth:
		s.navigateReauth(r, webSessionID, result)
	case authflow.FlowTypeAccountRecovery:
		s.navigateAccountRecovery(r, webSessionID, result)
	default:
		panic(fmt.Errorf("unexpected flow type: %v", s.StateTokenFlowResponse.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateSignup(r *http.Request, webSessionID string, result *Result) {
	s.navigateSignupPromote(r, webSessionID, result, AuthflowRouteSignup)
}

func (s *AuthflowScreenWithFlowResponse) navigatePromote(r *http.Request, webSessionID string, result *Result) {
	s.navigateSignupPromote(r, webSessionID, result, AuthflowRoutePromote)
}

func (s *AuthflowScreenWithFlowResponse) navigateSignupPromote(r *http.Request, webSessionID string, result *Result, expectedPath string) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		s.navigateStepIdentify(r, webSessionID, result, expectedPath)
	case config.AuthenticationFlowStepTypeCreateAuthenticator:
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.IntentSignupFlowStepCreateAuthenticatorData).Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			s.advance(AuthflowRouteCreatePassword, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			switch data := s.StateTokenFlowResponse.Action.Data.(type) {
			case declarative.NodeVerifyClaimData:
				// 1. We do not need to enter the target.
				switch data.OTPForm {
				case otp.FormCode:
					s.advance(AuthflowRouteEnterOOBOTP, result)
				case otp.FormLink:
					s.advance(AuthflowRouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
				}
			case declarative.IntentSignupFlowStepCreateAuthenticatorData:
				// 2. We need to enter the target.
				s.advance(AuthflowRouteSetupOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.advance(AuthflowRouteSetupTOTP, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			switch s.StateTokenFlowResponse.Action.Data.(type) {
			case declarative.NodeVerifyClaimData:
				// 1. We do not need to enter the target.
				channel := s.Screen.TakenChannel
				switch channel {
				case model.AuthenticatorOOBChannelSMS:
					s.advance(AuthflowRouteEnterOOBOTP, result)
				case model.AuthenticatorOOBChannelWhatsapp:
					s.advance(AuthflowRouteWhatsappOTP, result)
				default:
					panic(fmt.Errorf("unexpected channel: %v", channel))
				}
			case declarative.IntentSignupFlowStepCreateAuthenticatorData:
				// 2. We need to enter the target.
				s.advance(AuthflowRouteSetupOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case config.AuthenticationFlowStepTypeVerify:
		channel := s.Screen.TakenChannel
		data := s.StateTokenFlowResponse.Action.Data.(declarative.NodeVerifyClaimData)
		switch data.OTPForm {
		case otp.FormCode:
			switch channel {
			case model.AuthenticatorOOBChannelEmail:
				s.advance(AuthflowRouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelSMS:
				s.advance(AuthflowRouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.advance(AuthflowRouteWhatsappOTP, result)
			case "":
				// Verify may not have branches.
				s.advance(AuthflowRouteEnterOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case otp.FormLink:
			s.advance(AuthflowRouteOOBOTPLink, result)
		}
	case config.AuthenticationFlowStepTypeFillInUserProfile:
		panic(fmt.Errorf("fill_in_user_profile is not supported yet"))
	case config.AuthenticationFlowStepTypeViewRecoveryCode:
		s.advance(AuthflowRouteViewRecoveryCode, result)
	case config.AuthenticationFlowStepTypePromptCreatePasskey:
		s.advance(AuthflowRoutePromptCreatePasskey, result)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateLogin(r *http.Request, webSessionID string, result *Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		s.navigateStepIdentify(r, webSessionID, result, AuthflowRouteLogin)
	case config.AuthenticationFlowStepTypeAuthenticate:
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.IntentLoginFlowStepAuthenticateData).Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			s.advance(AuthflowRouteEnterPassword, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			data := s.StateTokenFlowResponse.Action.Data.(declarative.NodeVerifyClaimData)
			switch data.OTPForm {
			case otp.FormCode:
				s.advance(AuthflowRouteEnterOOBOTP, result)
			case otp.FormLink:
				s.advance(AuthflowRouteOOBOTPLink, result)
			default:
				panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
			}
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.advance(AuthflowRouteEnterTOTP, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			channel := s.Screen.TakenChannel
			switch channel {
			case model.AuthenticatorOOBChannelSMS:
				s.advance(AuthflowRouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.advance(AuthflowRouteWhatsappOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			s.advance(AuthflowRouteEnterRecoveryCode, result)
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			s.advance(AuthflowRouteUsePasskey, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case config.AuthenticationFlowStepTypeCheckAccountStatus:
		s.advance(AuthflowRouteAccountStatus, result)
	case config.AuthenticationFlowStepTypeTerminateOtherSessions:
		s.advance(AuthflowRouteTerminateOtherSessions, result)
	case config.AuthenticationFlowStepTypeChangePassword:
		s.advance(AuthflowRouteChangePassword, result)
	case config.AuthenticationFlowStepTypePromptCreatePasskey:
		s.advance(AuthflowRoutePromptCreatePasskey, result)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateReauth(r *http.Request, webSessionID string, result *Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		s.navigateStepIdentify(r, webSessionID, result, AuthflowRouteReauth)
	case config.AuthenticationFlowStepTypeAuthenticate:
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.IntentReauthFlowStepAuthenticateData).Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			s.advance(AuthflowRouteEnterPassword, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			data := s.StateTokenFlowResponse.Action.Data.(declarative.NodeVerifyClaimData)
			switch data.OTPForm {
			case otp.FormCode:
				s.advance(AuthflowRouteEnterOOBOTP, result)
			case otp.FormLink:
				s.advance(AuthflowRouteOOBOTPLink, result)
			default:
				panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
			}
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.advance(AuthflowRouteEnterTOTP, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			channel := s.Screen.TakenChannel
			switch channel {
			case model.AuthenticatorOOBChannelSMS:
				s.advance(AuthflowRouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.advance(AuthflowRouteWhatsappOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			s.advance(AuthflowRouteEnterRecoveryCode, result)
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			s.advance(AuthflowRouteUsePasskey, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateSignupLogin(r *http.Request, webSessionID string, result *Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		s.navigateStepIdentify(r, webSessionID, result, AuthflowRouteSignupLogin)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateAccountRecovery(r *http.Request, webSessionID string, result *Result) {
	navigate := func(path string) {
		u := *r.URL
		u.Path = path
		q := u.Query()
		q.Set(AuthflowQueryKey, s.Screen.StateToken.XStep)
		u.RawQuery = q.Encode()
		result.NavigationAction = "replace"
		result.RedirectURI = u.String()
	}
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		navigate(AuthflowRouteForgotPassword)
	case config.AuthenticationFlowStepTypeVerifyAccountRecoveryCode:
		navigate(AuthflowRouteForgotPasswordSuccess)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) advance(p string, result *Result) {
	q := url.Values{}
	q.Set(AuthflowQueryKey, s.Screen.StateToken.XStep)
	u, _ := url.Parse(p)
	u.RawQuery = q.Encode()

	result.NavigationAction = "advance"
	result.RedirectURI = u.String()
}

func (s *AuthflowScreenWithFlowResponse) navigateStepIdentify(r *http.Request, webSessionID string, result *Result, expectedPath string) {
	identification := s.StateTokenFlowResponse.Action.Identification
	switch identification {
	case "":
		fallthrough
	case config.AuthenticationFlowIdentificationIDToken:
		fallthrough
	case config.AuthenticationFlowIdentificationEmail:
		fallthrough
	case config.AuthenticationFlowIdentificationPhone:
		fallthrough
	case config.AuthenticationFlowIdentificationUsername:
		fallthrough
	case config.AuthenticationFlowIdentificationPasskey:
		// Redirect to the expected path with x_step set.
		u := *r.URL
		u.Path = expectedPath
		q := u.Query()
		q.Set(AuthflowQueryKey, s.Screen.StateToken.XStep)
		u.RawQuery = q.Encode()

		result.NavigationAction = "replace"
		result.RedirectURI = u.String()
	case config.AuthenticationFlowIdentificationOAuth:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.OAuthData)

		switch data.OAuthProviderType {
		case config.OAuthSSOProviderTypeWechat:
			s.advance(AuthflowRouteWechat, result)
		default:
			authorizationURL, _ := url.Parse(data.OAuthAuthorizationURL)
			q := authorizationURL.Query()

			state := AuthflowOAuthState{
				WebSessionID:     webSessionID,
				XStep:            s.Screen.StateToken.XStep,
				ErrorRedirectURI: expectedPath,
			}

			q.Set("state", state.Encode())
			authorizationURL.RawQuery = q.Encode()

			result.NavigationAction = "redirect"
			result.RedirectURI = authorizationURL.String()
		}

	default:
		panic(fmt.Errorf("unexpected identification: %v", identification))
	}
}
