package webapp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authflowclient"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
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

// Authflow remembers all seen screens. The screens could come from more than 1 flow.
// We intentionally DO NOT clear screens when a different flow is created.
// As long as the browser has a reference to x_step, a screen can be retrieved.
// This design is important to ensure traversing browser history will not cause flow not found error.
// See https://github.com/authgear/authgear-server/issues/3452
type Authflow struct {
	// AllScreens is x_step => screen.
	AllScreens map[string]*AuthflowScreen `json:"all_screens,omitempty"`
}

func (s *Session) RememberScreen(screen *AuthflowScreenWithFlowResponse) {
	if s.Authflow == nil {
		s.Authflow = &Authflow{}
	}

	if s.Authflow.AllScreens == nil {
		s.Authflow.AllScreens = make(map[string]*AuthflowScreen)
	}

	s.Authflow.AllScreens[screen.Screen.StateToken.XStep] = screen.Screen
}

// AuthflowStateToken pairs x_step with its underlying state_token.
type AuthflowStateToken struct {
	XStep      string `json:"x_step"`
	StateToken string `json:"state_token"`
}

func NewAuthflowStateToken(flowResponse *authflowclient.FlowResponse) *AuthflowStateToken {
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

func newAuthflowScreen(flowResponse *authflowclient.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	switch {
	case flowResponse.Action.Type == authflowclient.FlowActionTypeFinished:
		state := NewAuthflowStateToken(flowResponse)
		screen := &AuthflowScreen{
			PreviousXStep: previousXStep,
			PreviousInput: previousInput,
			StateToken:    state,
		}
		return screen
	default:
		switch flowResponse.Type {
		case authflowclient.FlowTypeSignup:
			return newAuthflowScreenSignup(flowResponse, previousXStep, previousInput)
		case authflowclient.FlowTypePromote:
			return newAuthflowScreenPromote(flowResponse, previousXStep, previousInput)
		case authflowclient.FlowTypeLogin:
			return newAuthflowScreenLogin(flowResponse, previousXStep, previousInput)
		case authflowclient.FlowTypeSignupLogin:
			return newAuthflowScreenSignupLogin(flowResponse, previousXStep, previousInput)
		case authflowclient.FlowTypeReauth:
			return newAuthflowScreenReauth(flowResponse, previousXStep, previousInput)
		case authflowclient.FlowTypeAccountRecovery:
			return newAuthflowScreenAccountRecovery(flowResponse, previousXStep, previousInput)
		default:
			panic(fmt.Errorf("unexpected flow type: %v", flowResponse.Type))
		}
	}
}

func newAuthflowScreenSignup(flowResponse *authflowclient.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	return newAuthflowScreenSignupPromote(flowResponse, previousXStep, previousInput)
}

func newAuthflowScreenPromote(flowResponse *authflowclient.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	return newAuthflowScreenSignupPromote(flowResponse, previousXStep, previousInput)
}

func newAuthflowScreenSignupPromote(flowResponse *authflowclient.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	state := NewAuthflowStateToken(flowResponse)
	screen := &AuthflowScreen{
		PreviousXStep: previousXStep,
		PreviousInput: previousInput,
		StateToken:    state,
	}
	switch flowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		// identify contains branches.
		screen.BranchStateToken = state
	case authflowclient.FlowActionTypeCreateAuthenticator:
		// authentication contains branches.
		screen.BranchStateToken = state
	case authflowclient.FlowActionTypeVerify:
		// verify MAY contain branches.
		if _, ok := authflowclient.CastDataChannels(flowResponse.Action.Data); ok {
			screen.BranchStateToken = state
		}
	case authflowclient.FlowActionTypeFillInUserProfile:
		// fill_in_user_profile contains NO branches.
		break
	case authflowclient.FlowActionTypeViewRecoveryCode:
		// view_recovery_code contains NO branches.
		break
	case authflowclient.FlowActionTypePromptCreatePasskey:
		// prompt_create_passkey contains NO branches.
		break
	default:
		panic(fmt.Errorf("unexpected action type: %v", flowResponse.Action.Type))
	}

	return screen
}

func newAuthflowScreenLogin(flowResponse *authflowclient.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	state := NewAuthflowStateToken(flowResponse)
	screen := &AuthflowScreen{
		PreviousXStep: previousXStep,
		PreviousInput: previousInput,
		StateToken:    state,
	}

	switch flowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		// identify contains branches.
		screen.BranchStateToken = state
	case authflowclient.FlowActionTypeAuthenticate:
		// authenticate contains branches.
		screen.BranchStateToken = state
	case authflowclient.FlowActionTypeCheckAccountStatus:
		// check_account_status contains NO branches.
		break
	case authflowclient.FlowActionTypeTerminateOtherSessions:
		// terminate_other_sessions contains NO branches.
		break
	case authflowclient.FlowActionTypeChangePassword:
		// change_password contains NO branches.
		break
	case authflowclient.FlowActionTypePromptCreatePasskey:
		// prompt_create_passkey contains NO branches.
		break
	default:
		panic(fmt.Errorf("unexpected action type: %v", flowResponse.Action.Type))
	}

	return screen
}

func newAuthflowScreenReauth(flowResponse *authflowclient.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	state := NewAuthflowStateToken(flowResponse)
	screen := &AuthflowScreen{
		PreviousXStep: previousXStep,
		PreviousInput: previousInput,
		StateToken:    state,
	}

	switch flowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		// identify contains branches.
		screen.BranchStateToken = state
	case authflowclient.FlowActionTypeAuthenticate:
		// authenticate contains branches.
		screen.BranchStateToken = state
	default:
		panic(fmt.Errorf("unexpected action type: %v", flowResponse.Action.Type))
	}

	return screen
}

func newAuthflowScreenSignupLogin(flowResponse *authflowclient.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	state := NewAuthflowStateToken(flowResponse)
	screen := &AuthflowScreen{
		PreviousXStep: previousXStep,
		PreviousInput: previousInput,
		StateToken:    state,
	}

	switch flowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		// identify contains branches.
		screen.BranchStateToken = state
	default:
		panic(fmt.Errorf("unexpected action type: %v", flowResponse.Action.Type))
	}

	return screen
}

func newAuthflowScreenAccountRecovery(flowResponse *authflowclient.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	state := NewAuthflowStateToken(flowResponse)
	screen := &AuthflowScreen{
		PreviousXStep: previousXStep,
		PreviousInput: previousInput,
		StateToken:    state,
	}

	switch flowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		// identify contains branches.
		screen.BranchStateToken = state
	case authflowclient.FlowActionTypeSelectDestination:
		// select_destination contains branches.
		screen.BranchStateToken = state
	case authflowclient.FlowActionTypeVerifyAccountRecoveryCode:
		// verify_account_recovery_code contains NO branches.
		break
	case authflowclient.FlowActionTypeResetPassword:
		// reset_password contains NO branches.
		break
	default:
		panic(fmt.Errorf("unexpected action type: %v", flowResponse.Action.Type))
	}

	return screen
}

type AuthflowScreenWithFlowResponse struct {
	Screen                       *AuthflowScreen
	StateTokenFlowResponse       *authflowclient.FlowResponse
	BranchStateTokenFlowResponse *authflowclient.FlowResponse
}

func NewAuthflowScreenWithFlowResponse(flowResponse *authflowclient.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreenWithFlowResponse {
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

func UpdateAuthflowScreenWithFlowResponse(screen *AuthflowScreenWithFlowResponse, flowResponse *authflowclient.FlowResponse) *AuthflowScreenWithFlowResponse {
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
	Input                 map[string]interface{}
	NewAuthflowScreenFull func(flowResponse *authflowclient.FlowResponse) *AuthflowScreenWithFlowResponse
}

func (TakeBranchResultInput) takeBranchResult() {}

func (s *AuthflowScreenWithFlowResponse) TakeBranch(index int, channel model.AuthenticatorOOBChannel) TakeBranchResult {
	switch s.StateTokenFlowResponse.Type {
	case authflowclient.FlowTypeSignup:
		return s.takeBranchSignup(index, channel)
	case authflowclient.FlowTypePromote:
		return s.takeBranchPromote(index, channel)
	case authflowclient.FlowTypeLogin:
		return s.takeBranchLogin(index, channel)
	case authflowclient.FlowTypeSignupLogin:
		return s.takeBranchSignupLogin(index)
	case authflowclient.FlowTypeReauth:
		return s.takeBranchReauth(index, channel)
	case authflowclient.FlowTypeAccountRecovery:
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
	switch s.StateTokenFlowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index)
	case authflowclient.FlowActionTypeCreateAuthenticator:
		var data authflowclient.DataCreateAuthenticator
		err := authflowclient.Cast(s.StateTokenFlowResponse.Action.Data, &data)
		if err != nil {
			panic(err)
		}
		option := data.Options[index]
		switch option.Authentication {
		case authflowclient.AuthenticationPrimaryPassword:
			fallthrough
		case authflowclient.AuthenticationSecondaryPassword:
			// Password branches can be taken by setting index.
			return s.takeBranchResultSimple(index)
		case authflowclient.AuthenticationSecondaryTOTP:
			// This branch requires input to take.
			input := map[string]interface{}{
				"authentication": "secondary_totp",
			}
			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflowclient.FlowResponse) *AuthflowScreenWithFlowResponse {
					var emptyChannel model.AuthenticatorOOBChannel
					isContinuation := func(flowResponse *authflowclient.FlowResponse) bool {
						return flowResponse.Action.Type == authflowclient.FlowActionTypeCreateAuthenticator &&
							flowResponse.Action.Authentication == authflowclient.AuthenticationSecondaryTOTP
					}

					return s.makeScreenForTakenBranch(flowResponse, input, &index, emptyChannel, isContinuation)
				},
			}
		case authflowclient.AuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case authflowclient.AuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPSMS:
			if channel == "" {
				channel = option.Channels[0]
			}
			input := map[string]interface{}{
				"authentication": option.Authentication,
				"channel":        channel,
			}
			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflowclient.FlowResponse) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflowclient.FlowResponse) bool {
						return flowResponse.Action.Type == authflowclient.FlowActionTypeCreateAuthenticator &&
							flowResponse.Action.Authentication == option.Authentication
					}

					return s.makeScreenForTakenBranch(flowResponse, input, &index, channel, isContinuation)
				},
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case authflowclient.FlowActionTypeVerify:
		// If we ever reach here, this means we have to choose channels.
		var data authflowclient.DataChannels
		err := authflowclient.Cast(s.StateTokenFlowResponse.Action.Data, &data)
		if err != nil {
			panic(err)
		}

		if channel == "" {
			channel = data.Channels[0]
		}
		input := map[string]interface{}{
			"channel": channel,
		}
		return TakeBranchResultInput{
			Input: input,
			NewAuthflowScreenFull: func(flowResponse *authflowclient.FlowResponse) *AuthflowScreenWithFlowResponse {
				var nilIndex *int
				isContinuation := func(flowResponse *authflowclient.FlowResponse) bool {
					return flowResponse.Action.Type == authflowclient.FlowActionTypeVerify
				}
				return s.makeScreenForTakenBranch(flowResponse, input, nilIndex, channel, isContinuation)
			},
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchLogin(index int, channel model.AuthenticatorOOBChannel) TakeBranchResult {
	switch s.StateTokenFlowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index)
	case authflowclient.FlowActionTypeAuthenticate:
		var data authflowclient.DataAuthenticate
		err := authflowclient.Cast(s.StateTokenFlowResponse.Action.Data, &data)
		if err != nil {
			panic(err)
		}
		option := data.Options[index]
		switch option.Authentication {
		case authflowclient.AuthenticationPrimaryPassword:
			fallthrough
		case authflowclient.AuthenticationSecondaryPassword:
			fallthrough
		case authflowclient.AuthenticationSecondaryTOTP:
			fallthrough
		case authflowclient.AuthenticationRecoveryCode:
			fallthrough
		case authflowclient.AuthenticationPrimaryPasskey:
			// All these can take the branch simply by setting index.
			return s.takeBranchResultSimple(index)
		case authflowclient.AuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case authflowclient.AuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPSMS:
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
				NewAuthflowScreenFull: func(flowResponse *authflowclient.FlowResponse) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflowclient.FlowResponse) bool {
						return flowResponse.Action.Type == authflowclient.FlowActionTypeAuthenticate && flowResponse.Action.Authentication == option.Authentication
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
	switch s.StateTokenFlowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		// In identify, id_token is used.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index)
	case authflowclient.FlowActionTypeAuthenticate:
		var data authflowclient.DataAuthenticate
		err := authflowclient.Cast(s.StateTokenFlowResponse.Action.Data, &data)
		if err != nil {
			panic(err)
		}
		option := data.Options[index]
		switch option.Authentication {
		case authflowclient.AuthenticationPrimaryPassword:
			fallthrough
		case authflowclient.AuthenticationSecondaryPassword:
			fallthrough
		case authflowclient.AuthenticationSecondaryTOTP:
			fallthrough
		case authflowclient.AuthenticationPrimaryPasskey:
			// All these can take the branch simply by setting index.
			return s.takeBranchResultSimple(index)
		case authflowclient.AuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case authflowclient.AuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPSMS:
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
				NewAuthflowScreenFull: func(flowResponse *authflowclient.FlowResponse) *AuthflowScreenWithFlowResponse {
					isContinuation := func(flowResponse *authflowclient.FlowResponse) bool {
						return flowResponse.Action.Type == authflowclient.FlowActionTypeAuthenticate && flowResponse.Action.Authentication == option.Authentication
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
	switch s.StateTokenFlowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchAccountRecovery(index int) TakeBranchResult {
	switch s.StateTokenFlowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
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
	flowResponse *authflowclient.FlowResponse,
	input map[string]interface{},
	index *int,
	channel model.AuthenticatorOOBChannel,
	isContinuation func(flowResponse *authflowclient.FlowResponse) bool,
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
	case authflowclient.FlowTypeSignup:
		s.navigateSignup(r, webSessionID, result)
	case authflowclient.FlowTypePromote:
		s.navigatePromote(r, webSessionID, result)
	case authflowclient.FlowTypeLogin:
		s.navigateLogin(r, webSessionID, result)
	case authflowclient.FlowTypeSignupLogin:
		s.navigateSignupLogin(r, webSessionID, result)
	case authflowclient.FlowTypeReauth:
		s.navigateReauth(r, webSessionID, result)
	case authflowclient.FlowTypeAccountRecovery:
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

// nolint: gocyclo
func (s *AuthflowScreenWithFlowResponse) navigateSignupPromote(r *http.Request, webSessionID string, result *Result, expectedPath string) {
	switch s.StateTokenFlowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		s.navigateStepIdentify(r, webSessionID, result, expectedPath)
	case authflowclient.FlowActionTypeCreateAuthenticator:
		var data authflowclient.DataCreateAuthenticator
		err := authflowclient.Cast(s.BranchStateTokenFlowResponse.Action.Data, &data)
		if err != nil {
			panic(err)
		}
		options := data.Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case authflowclient.AuthenticationPrimaryPassword:
			fallthrough
		case authflowclient.AuthenticationSecondaryPassword:
			s.advance(AuthflowRouteCreatePassword, result)
		case authflowclient.AuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPEmail:
			dataVerifyClaim, dataCreateAuthenticator, err := authflowclient.CastVerifyClaimOrCreateAuthenticator(s.StateTokenFlowResponse.Action.Data)
			if err != nil {
				panic(err)
			}

			switch {
			case dataVerifyClaim != nil:
				// 1. We do not need to enter the target.
				switch dataVerifyClaim.OTPForm {
				case otp.FormCode:
					s.advance(AuthflowRouteEnterOOBOTP, result)
				case otp.FormLink:
					s.advance(AuthflowRouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", dataVerifyClaim.OTPForm))
				}
			case dataCreateAuthenticator != nil:
				// 2. We need to enter the target.
				s.advance(AuthflowRouteSetupOOBOTP, result)
			}
		case authflowclient.AuthenticationSecondaryTOTP:
			s.advance(AuthflowRouteSetupTOTP, result)
		case authflowclient.AuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPSMS:
			dataVerifyClaim, dataCreateAuthenticator, err := authflowclient.CastVerifyClaimOrCreateAuthenticator(s.StateTokenFlowResponse.Action.Data)
			if err != nil {
				panic(err)
			}

			switch {
			case dataVerifyClaim != nil:
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
			case dataCreateAuthenticator != nil:
				// 2. We need to enter the target.
				s.advance(AuthflowRouteSetupOOBOTP, result)
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case authflowclient.FlowActionTypeVerify:
		channel := s.Screen.TakenChannel
		var data authflowclient.DataVerifyClaim
		err := authflowclient.Cast(s.StateTokenFlowResponse.Action.Data, &data)
		if err != nil {
			panic(err)
		}
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
	case authflowclient.FlowActionTypeFillInUserProfile:
		panic(fmt.Errorf("fill_in_user_profile is not supported yet"))
	case authflowclient.FlowActionTypeViewRecoveryCode:
		s.advance(AuthflowRouteViewRecoveryCode, result)
	case authflowclient.FlowActionTypePromptCreatePasskey:
		s.advance(AuthflowRoutePromptCreatePasskey, result)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateLogin(r *http.Request, webSessionID string, result *Result) {
	switch s.StateTokenFlowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		s.navigateStepIdentify(r, webSessionID, result, AuthflowRouteLogin)
	case authflowclient.FlowActionTypeAuthenticate:
		var data authflowclient.DataAuthenticate
		err := authflowclient.Cast(s.BranchStateTokenFlowResponse.Action.Data, &data)
		if err != nil {
			panic(err)
		}
		options := data.Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case authflowclient.AuthenticationPrimaryPassword:
			fallthrough
		case authflowclient.AuthenticationSecondaryPassword:
			s.advance(AuthflowRouteEnterPassword, result)
		case authflowclient.AuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPEmail:
			var data authflowclient.DataVerifyClaim
			err := authflowclient.Cast(s.StateTokenFlowResponse.Action.Data, &data)
			if err != nil {
				panic(err)
			}
			switch data.OTPForm {
			case otp.FormCode:
				s.advance(AuthflowRouteEnterOOBOTP, result)
			case otp.FormLink:
				s.advance(AuthflowRouteOOBOTPLink, result)
			default:
				panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
			}
		case authflowclient.AuthenticationSecondaryTOTP:
			s.advance(AuthflowRouteEnterTOTP, result)
		case authflowclient.AuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPSMS:
			channel := s.Screen.TakenChannel
			switch channel {
			case model.AuthenticatorOOBChannelSMS:
				s.advance(AuthflowRouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.advance(AuthflowRouteWhatsappOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case authflowclient.AuthenticationRecoveryCode:
			s.advance(AuthflowRouteEnterRecoveryCode, result)
		case authflowclient.AuthenticationPrimaryPasskey:
			s.advance(AuthflowRouteUsePasskey, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case authflowclient.FlowActionTypeCheckAccountStatus:
		s.advance(AuthflowRouteAccountStatus, result)
	case authflowclient.FlowActionTypeTerminateOtherSessions:
		s.advance(AuthflowRouteTerminateOtherSessions, result)
	case authflowclient.FlowActionTypeChangePassword:
		s.advance(AuthflowRouteChangePassword, result)
	case authflowclient.FlowActionTypePromptCreatePasskey:
		s.advance(AuthflowRoutePromptCreatePasskey, result)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateReauth(r *http.Request, webSessionID string, result *Result) {
	switch s.StateTokenFlowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		s.navigateStepIdentify(r, webSessionID, result, AuthflowRouteReauth)
	case authflowclient.FlowActionTypeAuthenticate:
		var data authflowclient.DataAuthenticate
		err := authflowclient.Cast(s.BranchStateTokenFlowResponse.Action.Data, &data)
		if err != nil {
			panic(err)
		}
		options := data.Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case authflowclient.AuthenticationPrimaryPassword:
			fallthrough
		case authflowclient.AuthenticationSecondaryPassword:
			s.advance(AuthflowRouteEnterPassword, result)
		case authflowclient.AuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPEmail:
			var data authflowclient.DataVerifyClaim
			err := authflowclient.Cast(s.StateTokenFlowResponse.Action.Data, &data)
			if err != nil {
				panic(err)
			}
			switch data.OTPForm {
			case otp.FormCode:
				s.advance(AuthflowRouteEnterOOBOTP, result)
			case otp.FormLink:
				s.advance(AuthflowRouteOOBOTPLink, result)
			default:
				panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
			}
		case authflowclient.AuthenticationSecondaryTOTP:
			s.advance(AuthflowRouteEnterTOTP, result)
		case authflowclient.AuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case authflowclient.AuthenticationSecondaryOOBOTPSMS:
			channel := s.Screen.TakenChannel
			switch channel {
			case model.AuthenticatorOOBChannelSMS:
				s.advance(AuthflowRouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.advance(AuthflowRouteWhatsappOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case authflowclient.AuthenticationPrimaryPasskey:
			s.advance(AuthflowRouteUsePasskey, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateSignupLogin(r *http.Request, webSessionID string, result *Result) {
	switch s.StateTokenFlowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
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
	switch s.StateTokenFlowResponse.Action.Type {
	case authflowclient.FlowActionTypeIdentify:
		navigate(AuthflowRouteForgotPassword)
	case authflowclient.FlowActionTypeVerifyAccountRecoveryCode:
		navigate(AuthflowRouteForgotPasswordSuccess)
	case authflowclient.FlowActionTypeResetPassword:
		navigate(AuthflowRouteResetPassword)
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
	case authflowclient.IdentificationIDToken:
		fallthrough
	case authflowclient.IdentificationEmail:
		fallthrough
	case authflowclient.IdentificationPhone:
		fallthrough
	case authflowclient.IdentificationUsername:
		fallthrough
	case authflowclient.IdentificationPasskey:
		// Redirect to the expected path with x_step set.
		u := *r.URL
		u.Path = expectedPath
		q := u.Query()
		q.Set(AuthflowQueryKey, s.Screen.StateToken.XStep)
		u.RawQuery = q.Encode()

		result.NavigationAction = "replace"
		result.RedirectURI = u.String()
	case authflowclient.IdentificationOAuth:
		var data authflowclient.DataOAuth
		err := authflowclient.Cast(s.StateTokenFlowResponse.Action.Data, &data)
		if err != nil {
			panic(err)
		}

		switch data.OAuthProviderType {
		case "wechat":
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

func DeriveAuthflowFinishPath(response *authflowclient.FlowResponse) string {
	switch response.Type {
	case authflowclient.FlowTypeAccountRecovery:
		return AuthflowRouteResetPasswordSuccess
	}
	return ""
}
