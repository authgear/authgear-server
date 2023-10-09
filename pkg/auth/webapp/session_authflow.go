package webapp

import (
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

const AuthflowQueryKey = "x_step"

// Authflow stores the necessary information for webapp to run an authflow, including
// navigation, and branching.
type Authflow struct {
	// FlowID is the ID of the authflow.
	FlowID string `json:"flow_id"`
	// AllScreens is x_step => screen.
	AllScreens map[string]*AuthflowScreen `json:"all_screens,omitempty"`
}

func NewAuthflow(flowID string, initialScreen *AuthflowScreenWithFlowResponse) *Authflow {
	af := &Authflow{
		FlowID:     flowID,
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
}

func newAuthflowScreen(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
	switch flowResponse.Type {
	case authflow.FlowTypeSignup:
		return newAuthflowScreenSignup(flowResponse, previousXStep, previousInput)
	case authflow.FlowTypeLogin:
		return newAuthflowScreenLogin(flowResponse, previousXStep, previousInput)
	case authflow.FlowTypeSignupLogin:
		return newAuthflowScreenSignupLogin(flowResponse, previousXStep, previousInput)
	default:
		panic(fmt.Errorf("unexpected flow type: %v", flowResponse.Type))
	}
}

func newAuthflowScreenSignup(flowResponse *authflow.FlowResponse, previousXStep string, previousInput map[string]interface{}) *AuthflowScreen {
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
		// authentication contains branches.
		screen.BranchStateToken = state
	case config.AuthenticationFlowStepTypeVerify:
		// verify MAY contain branches.
		if _, ok := flowResponse.Action.Data.(declarative.IntentVerifyClaimData); ok {
			screen.BranchStateToken = state
		}
	case config.AuthenticationFlowStepTypeUserProfile:
		// user_profile contains NO branches.
		break
	case config.AuthenticationFlowStepTypeRecoveryCode:
		// recovery_code contains NO branches.
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
	case authflow.FlowTypeLogin:
		return s.takeBranchLogin(index, channel)
	case authflow.FlowTypeSignupLogin:
		return s.takeBranchSignupLogin(index)
	default:
		panic(fmt.Errorf("unexpected flow type: %v", s.StateTokenFlowResponse.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) takeBranchSignup(index int, channel model.AuthenticatorOOBChannel) TakeBranchResult {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		// In identify, the user input actually takes the branch.
		// The branch taken here is unimportant.
		return s.takeBranchResultSimple(index)
	case config.AuthenticationFlowStepTypeAuthenticate:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.IntentSignupFlowStepAuthenticateData)
		option := data.Options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			// All these can take the branch simply by setting index.
			return s.takeBranchResultSimple(index)
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			// This branch requires input to take.
			input := map[string]interface{}{
				"authentication": "secondary_totp",
			}
			return TakeBranchResultInput{
				Input: input,
				NewAuthflowScreenFull: func(flowResponse *authflow.FlowResponse) *AuthflowScreenWithFlowResponse {
					xStep := s.Screen.StateToken.XStep
					screen := NewAuthflowScreenWithFlowResponse(flowResponse, xStep, input)
					screen.Screen.BranchStateToken = s.Screen.StateToken
					screen.BranchStateTokenFlowResponse = s.StateTokenFlowResponse
					screen.Screen.TakenBranchIndex = &index
					return screen
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
				xStep := s.Screen.StateToken.XStep
				screen := NewAuthflowScreenWithFlowResponse(flowResponse, xStep, input)
				screen.Screen.BranchStateToken = s.Screen.StateToken
				screen.BranchStateTokenFlowResponse = s.StateTokenFlowResponse
				screen.Screen.TakenChannel = channel
				return screen
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
					xStep := s.Screen.StateToken.XStep
					screen := NewAuthflowScreenWithFlowResponse(flowResponse, xStep, input)
					screen.Screen.BranchStateToken = s.Screen.StateToken
					screen.BranchStateTokenFlowResponse = s.StateTokenFlowResponse
					screen.Screen.TakenBranchIndex = &index
					screen.Screen.TakenChannel = channel
					return screen
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

func (s *AuthflowScreenWithFlowResponse) Navigate(r *http.Request, result *Result) {
	if s.HasBranchToTake() {
		panic(fmt.Errorf("expected screen to have its branches taken"))
	}

	switch s.StateTokenFlowResponse.Type {
	case authflow.FlowTypeSignup:
		s.navigateSignup(r, result)
	case authflow.FlowTypeLogin:
		s.navigateLogin(r, result)
	case authflow.FlowTypeSignupLogin:
		s.navigateSignupLogin(r, result)
	default:
		panic(fmt.Errorf("unexpected flow type: %v", s.StateTokenFlowResponse.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateSignup(r *http.Request, result *Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		s.navigateStepIdentify(r, result)
	case config.AuthenticationFlowStepTypeAuthenticate:
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.IntentSignupFlowStepAuthenticateData).Options
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
			s.advance(AuthflowRouteSetupTOTP, result)
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
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case config.AuthenticationFlowStepTypeVerify:
		channel := s.Screen.TakenChannel
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
	case config.AuthenticationFlowStepTypeUserProfile:
		panic(fmt.Errorf("user_profile is not supported yet"))
	case config.AuthenticationFlowStepTypeRecoveryCode:
		s.advance(AuthflowRouteViewRecoveryCode, result)
	case config.AuthenticationFlowStepTypePromptCreatePasskey:
		s.advance(AuthflowRoutePromptCreatePasskey, result)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateLogin(r *http.Request, result *Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		s.navigateStepIdentify(r, result)
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
	case config.AuthenticationFlowStepTypeChangePassword:
		s.advance(AuthflowRouteChangePassword, result)
	case config.AuthenticationFlowStepTypePromptCreatePasskey:
		s.advance(AuthflowRoutePromptCreatePasskey, result)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (s *AuthflowScreenWithFlowResponse) navigateSignupLogin(r *http.Request, result *Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		s.navigateStepIdentify(r, result)
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

func (s *AuthflowScreenWithFlowResponse) navigateStepIdentify(r *http.Request, result *Result) {
	identification := s.StateTokenFlowResponse.Action.Identification
	switch identification {
	case "":
		fallthrough
	case config.AuthenticationFlowIdentificationEmail:
		fallthrough
	case config.AuthenticationFlowIdentificationPhone:
		fallthrough
	case config.AuthenticationFlowIdentificationUsername:
		fallthrough
	case config.AuthenticationFlowIdentificationPasskey:
		// Stay in the same page with x_step set.
		u := *r.URL
		q := u.Query()
		q.Set(AuthflowQueryKey, s.Screen.StateToken.XStep)
		u.RawQuery = q.Encode()

		result.NavigationAction = "replace"
		result.RedirectURI = u.String()
	case config.AuthenticationFlowIdentificationOAuth:
		// Redirect to the external OAuth provider.
		var authorizationURLStr string
		switch data := s.StateTokenFlowResponse.Action.Data.(type) {
		case declarative.NodeOAuthData:
			authorizationURLStr = data.OAuthAuthorizationURL
		case declarative.NodeLookupIdentityOAuthData:
			authorizationURLStr = data.OAuthAuthorizationURL
		default:
			panic(fmt.Errorf("unexpected data type: %T", s.StateTokenFlowResponse.Action.Data))
		}

		authorizationURL, _ := url.Parse(authorizationURLStr)
		q := authorizationURL.Query()
		// Set state=<value of x_step> so that the frontend can resume.
		q.Set("state", s.Screen.StateToken.XStep)
		authorizationURL.RawQuery = q.Encode()

		result.NavigationAction = "redirect"
		result.RedirectURI = authorizationURL.String()
	default:
		panic(fmt.Errorf("unexpected identification: %v", identification))
	}
}
