package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
)

const (
	AuthflowRouteLogin   = "/login"
	AuthflowRouteSignup  = "/signup"
	AuthflowRoutePromote = "/flows/promote_user"
	AuthflowRouteReauth  = "/reauth"
	// AuthflowRouteSignupLogin is login because login page has passkey.
	AuthflowRouteSignupLogin = AuthflowRouteLogin

	AuthflowRouteTerminateOtherSessions = "/authflow/terminate_other_sessions"
	// nolint: gosec
	AuthflowRoutePromptCreatePasskey = "/authflow/prompt_create_passkey"
	AuthflowRouteViewRecoveryCode    = "/authflow/view_recovery_code"
	// nolint: gosec
	AuthflowRouteCreatePassword = "/authflow/create_password"
	// nolint: gosec
	AuthflowRouteChangePassword = "/authflow/change_password"
	// nolint: gosec
	AuthflowRouteEnterPassword     = "/authflow/enter_password"
	AuthflowRouteEnterRecoveryCode = "/authflow/enter_recovery_code"
	AuthflowRouteEnterOOBOTP       = "/authflow/enter_oob_otp"
	AuthflowRouteWhatsappOTP       = "/authflow/whatsapp_otp"
	AuthflowRouteOOBOTPLink        = "/authflow/oob_otp_link"
	AuthflowRouteEnterTOTP         = "/authflow/enter_totp"
	AuthflowRouteSetupTOTP         = "/authflow/setup_totp"
	AuthflowRouteSetupOOBOTP       = "/authflow/setup_oob_otp"
	// nolint: gosec
	AuthflowRouteUsePasskey = "/authflow/use_passkey"
	// nolint: gosec
	AuthflowRouteForgotPassword = "/authflow/forgot_password"
	// nolint: gosec
	AuthflowRouteForgotPasswordOTP = "/authflow/forgot_password/otp"
	// nolint: gosec
	AuthflowRouteForgotPasswordSuccess = "/authflow/forgot_password/success"
	// nolint: gosec
	AuthflowRouteResetPassword = "/authflow/reset_password"
	// nolint: gosec
	AuthflowRouteResetPasswordSuccess = "/authflow/reset_password/success"
	AuthflowRouteWechat               = "/authflow/wechat"

	// The following routes are dead ends.
	AuthflowRouteAccountStatus   = "/authflow/account_status"
	AuthflowRouteNoAuthenticator = "/authflow/no_authenticator"

	AuthflowRouteFinishFlow = "/authflow/finish"
)

type AuthflowNavigatorEndpointsProvider interface {
	ErrorEndpointURL(uiImpl config.UIImplementation) *url.URL
	SelectAccountEndpointURL(uiImpl config.UIImplementation) *url.URL
}

type AuthflowNavigator struct {
	Endpoints   AuthflowNavigatorEndpointsProvider
	UIConfig    *config.UIConfig
	ErrorCookie *ErrorCookie
}

func (n *AuthflowNavigator) NavigateNonRecoverableError(r *http.Request, u *url.URL, e error) {
	switch {
	case user.IsAccountStatusError(e):
		u.Path = AuthflowRouteAccountStatus
	case errors.Is(e, api.ErrNoAuthenticator):
		u.Path = AuthflowRouteNoAuthenticator
	case errors.Is(e, authflow.ErrFlowNotFound):
		u.Path = n.Endpoints.ErrorEndpointURL(n.UIConfig.Implementation).Path
	case apierrors.IsKind(e, WebUIInvalidSession):
		// Show WebUIInvalidSession error in different page.
		u.Path = n.Endpoints.ErrorEndpointURL(n.UIConfig.Implementation).Path
	case r.Method == http.MethodGet:
		// If the request method is Get, avoid redirect back to the same path
		// which causes infinite redirect loop
		u.Path = n.Endpoints.ErrorEndpointURL(n.UIConfig.Implementation).Path
	}
}

func (n *AuthflowNavigator) Navigate(s *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result) {
	if s.HasBranchToTake() {
		panic(fmt.Errorf("expected screen to have its branches taken"))
	}

	if s.StateTokenFlowResponse.Action.Type == authflow.FlowActionTypeFinished {
		s.Advance(AuthflowRouteFinishFlow, result)
		return
	}

	switch s.StateTokenFlowResponse.Type {
	case authflow.FlowTypeSignup:
		n.navigateSignup(s, r, webSessionID, result)
	case authflow.FlowTypePromote:
		n.navigatePromote(s, r, webSessionID, result)
	case authflow.FlowTypeLogin:
		n.navigateLogin(s, r, webSessionID, result)
	case authflow.FlowTypeSignupLogin:
		n.navigateSignupLogin(s, r, webSessionID, result)
	case authflow.FlowTypeReauth:
		n.navigateReauth(s, r, webSessionID, result)
	case authflow.FlowTypeAccountRecovery:
		n.navigateAccountRecovery(s, r, webSessionID, result)
	default:
		panic(fmt.Errorf("unexpected flow type: %v", s.StateTokenFlowResponse.Type))
	}
}

func (n *AuthflowNavigator) NavigateChangePasswordSuccessPage(s *AuthflowScreen, r *http.Request, webSessionID string) (result *Result) {
	return &Result{}
}

func (n *AuthflowNavigator) NavigateResetPasswordSuccessPage() string {
	return AuthflowRouteResetPasswordSuccess
}

func (n *AuthflowNavigator) navigateSignup(s *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result) {
	n.navigateSignupPromote(s, r, webSessionID, result, AuthflowRouteSignup)
}

func (n *AuthflowNavigator) navigatePromote(s *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result) {
	n.navigateSignupPromote(s, r, webSessionID, result, AuthflowRoutePromote)
}

//nolint:gocognit
func (n *AuthflowNavigator) navigateSignupPromote(s *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result, expectedPath string) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(s, r, webSessionID, result, expectedPath)
	case config.AuthenticationFlowStepTypeCreateAuthenticator:
		// If the current step already tells the authentication, use it
		authentication := s.StateTokenFlowResponse.Action.Authentication
		if authentication == "" {
			// Else, get it from the first option of the branch step
			options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.IntentSignupFlowStepCreateAuthenticatorData).Options
			index := *s.Screen.TakenBranchIndex
			option := options[index]
			authentication = option.Authentication
		}
		switch authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			s.Advance(AuthflowRouteCreatePassword, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			switch data := s.StateTokenFlowResponse.Action.Data.(type) {
			case declarative.VerifyOOBOTPData:
				// 1. We do not need to enter the target.
				switch data.OTPForm {
				case otp.FormCode:
					s.Advance(AuthflowRouteEnterOOBOTP, result)
				case otp.FormLink:
					s.Advance(AuthflowRouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
				}
			case declarative.IntentSignupFlowStepCreateAuthenticatorData:
				// 2. We need to enter the target.
				s.Advance(AuthflowRouteSetupOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.Advance(AuthflowRouteSetupTOTP, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			data := s.StateTokenFlowResponse.Action.Data
			switch data := data.(type) {
			case declarative.VerifyOOBOTPData:
				// 1. We do not need to enter the target.
				channel := data.Channel
				switch channel {
				case model.AuthenticatorOOBChannelSMS:
					s.Advance(AuthflowRouteEnterOOBOTP, result)
				case model.AuthenticatorOOBChannelWhatsapp:
					s.Advance(AuthflowRouteWhatsappOTP, result)
				default:
					panic(fmt.Errorf("unexpected channel: %v", channel))
				}
			case declarative.IntentSignupFlowStepCreateAuthenticatorData:
				// 2. We need to enter the target.
				s.Advance(AuthflowRouteSetupOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", s.StateTokenFlowResponse.Action.Authentication))
		}
	case config.AuthenticationFlowStepTypeVerify:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.VerifyOOBOTPData)
		channel := data.Channel
		switch data.OTPForm {
		case otp.FormCode:
			switch channel {
			case model.AuthenticatorOOBChannelEmail:
				s.Advance(AuthflowRouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelSMS:
				s.Advance(AuthflowRouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.Advance(AuthflowRouteWhatsappOTP, result)
			case "":
				// Verify may not have branches.
				s.Advance(AuthflowRouteEnterOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case otp.FormLink:
			s.Advance(AuthflowRouteOOBOTPLink, result)
		}
	case config.AuthenticationFlowStepTypeFillInUserProfile:
		panic(fmt.Errorf("fill_in_user_profile is not supported yet"))
	case config.AuthenticationFlowStepTypeViewRecoveryCode:
		s.Advance(AuthflowRouteViewRecoveryCode, result)
	case config.AuthenticationFlowStepTypePromptCreatePasskey:
		s.Advance(AuthflowRoutePromptCreatePasskey, result)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowNavigator) navigateStepIdentify(s *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result, expectedPath string) {
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
		case wechat.Type:
			s.Advance(AuthflowRouteWechat, result)
		default:
			authorizationURL, _ := url.Parse(data.OAuthAuthorizationURL)
			q := authorizationURL.Query()

			state := webappoauth.WebappOAuthState{
				WebSessionID:     webSessionID,
				UIImplementation: config.UIImplementationAuthflow,
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

func (n *AuthflowNavigator) navigateLogin(s *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(s, r, webSessionID, result, AuthflowRouteLogin)
	case config.AuthenticationFlowStepTypeAuthenticate:
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.StepAuthenticateData).Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			s.Advance(AuthflowRouteEnterPassword, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			switch data := s.StateTokenFlowResponse.Action.Data.(type) {
			case declarative.VerifyOOBOTPData:
				switch data.OTPForm {
				case otp.FormCode:
					s.Advance(AuthflowRouteEnterOOBOTP, result)
				case otp.FormLink:
					s.Advance(AuthflowRouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
				}
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.Advance(AuthflowRouteEnterTOTP, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			channel := s.Screen.TakenChannel
			switch channel {
			case model.AuthenticatorOOBChannelSMS:
				s.Advance(AuthflowRouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.Advance(AuthflowRouteWhatsappOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			s.Advance(AuthflowRouteEnterRecoveryCode, result)
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			s.Advance(AuthflowRouteUsePasskey, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case config.AuthenticationFlowStepTypeCheckAccountStatus:
		s.Advance(AuthflowRouteAccountStatus, result)
	case config.AuthenticationFlowStepTypeTerminateOtherSessions:
		s.Advance(AuthflowRouteTerminateOtherSessions, result)
	case config.AuthenticationFlowStepTypeChangePassword:
		s.Advance(AuthflowRouteChangePassword, result)
	case config.AuthenticationFlowStepTypePromptCreatePasskey:
		s.Advance(AuthflowRoutePromptCreatePasskey, result)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowNavigator) navigateReauth(s *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(s, r, webSessionID, result, AuthflowRouteReauth)
	case config.AuthenticationFlowStepTypeAuthenticate:
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.StepAuthenticateData).Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			s.Advance(AuthflowRouteEnterPassword, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			switch data := s.StateTokenFlowResponse.Action.Data.(type) {
			case declarative.VerifyOOBOTPData:
				switch data.OTPForm {
				case otp.FormCode:
					s.Advance(AuthflowRouteEnterOOBOTP, result)
				case otp.FormLink:
					s.Advance(AuthflowRouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
				}
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.Advance(AuthflowRouteEnterTOTP, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			channel := s.Screen.TakenChannel
			switch channel {
			case model.AuthenticatorOOBChannelSMS:
				s.Advance(AuthflowRouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.Advance(AuthflowRouteWhatsappOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			s.Advance(AuthflowRouteUsePasskey, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowNavigator) navigateSignupLogin(s *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(s, r, webSessionID, result, AuthflowRouteSignupLogin)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowNavigator) navigateAccountRecovery(s *AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *Result) {
	navigate := func(path string, query *url.Values) {
		u := *r.URL
		u.Path = path
		q := u.Query()
		q.Set(AuthflowQueryKey, s.Screen.StateToken.XStep)
		for k, param := range *query {
			for _, p := range param {
				q.Add(k, p)
			}
		}
		u.RawQuery = q.Encode()
		result.NavigationAction = "replace"
		result.RedirectURI = u.String()
	}
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		navigate(AuthflowRouteForgotPassword, &url.Values{})
	case config.AuthenticationFlowStepTypeSelectDestination:
		navigate(AuthflowRouteForgotPassword, &url.Values{})
	case config.AuthenticationFlowStepTypeVerifyAccountRecoveryCode:
		data, ok := s.StateTokenFlowResponse.Action.Data.(declarative.IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData)
		if ok && data.OTPForm == declarative.AccountRecoveryOTPFormCode {
			navigate(AuthflowRouteForgotPasswordOTP, &url.Values{"x_can_back_to_login": []string{"true"}})
		} else {
			navigate(AuthflowRouteForgotPasswordSuccess, &url.Values{"x_can_back_to_login": []string{"false"}})
		}
	case config.AuthenticationFlowStepTypeResetPassword:
		navigate(AuthflowRouteResetPassword, &url.Values{})
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowNavigator) NavigateSelectAccount(result *Result) {
	url := n.Endpoints.SelectAccountEndpointURL(n.UIConfig.Implementation)
	result.RedirectURI = url.String()
}
