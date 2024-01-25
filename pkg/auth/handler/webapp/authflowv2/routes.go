package authflowv2

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	handlerwebapp "github.com/authgear/authgear-server/pkg/auth/handler/webapp"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	AuthflowV2RouteLogin   = "/login"
	AuthflowV2RouteSignup  = "/signup"
	AuthflowV2RoutePromote = "/flows/promote_user"
	AuthflowV2RouteReauth  = "/reauth"
	// AuthflowV2RouteSignupLogin is login because login page has passkey.
	AuthflowV2RouteSignupLogin = AuthflowV2RouteLogin

	AuthflowV2RouteTerminateOtherSessions = "/authflow/v2/terminate_other_sessions"
	// nolint: gosec
	AuthflowV2RoutePromptCreatePasskey = "/authflow/v2/prompt_create_passkey"
	AuthflowV2RouteViewRecoveryCode    = "/authflow/v2/view_recovery_code"
	// nolint: gosec
	AuthflowV2RouteCreatePassword = "/authflow/v2/create_password"
	// nolint: gosec
	AuthflowV2RouteChangePassword = "/authflow/v2/change_password"
	// nolint: gosec
	AuthflowV2RouteEnterPassword     = "/authflow/v2/enter_password"
	AuthflowV2RouteEnterRecoveryCode = "/authflow/v2/enter_recovery_code"
	AuthflowV2RouteEnterOOBOTP       = "/authflow/v2/enter_oob_otp"
	AuthflowV2RouteWhatsappOTP       = "/authflow/v2/whatsapp_otp"
	AuthflowV2RouteOOBOTPLink        = "/authflow/v2/oob_otp_link"
	AuthflowV2RouteEnterTOTP         = "/authflow/v2/enter_totp"
	AuthflowV2RouteSetupTOTP         = "/authflow/v2/setup_totp"
	AuthflowV2RouteSetupOOBOTP       = "/authflow/v2/setup_oob_otp"
	// nolint: gosec
	AuthflowV2RouteUsePasskey = "/authflow/v2/use_passkey"
	// nolint: gosec
	AuthflowV2RouteForgotPassword = "/authflow/v2/forgot_password"
	// nolint: gosec
	AuthflowV2RouteForgotPasswordOTP = "/authflow/v2/forgot_password/otp"
	// nolint: gosec
	AuthflowV2RouteForgotPasswordSuccess = "/authflow/v2/forgot_password/success"
	// nolint: gosec
	AuthflowV2RouteResetPassword = "/authflow/v2/reset_password"
	// nolint: gosec
	AuthflowV2RouteResetPasswordSuccess = "/authflow/v2/reset_password/success"
	AuthflowV2RouteWechat               = "/authflow/v2/wechat"

	// The following routes are dead ends.
	AuthflowV2RouteAccountStatus   = "/authflow/v2/account_status"
	AuthflowV2RouteNoAuthenticator = "/authflow/v2/no_authenticator"
)

type AuthflowV2NavigatorEndpointsProvider interface {
	ErrorEndpointURL(uiImpl config.UIImplementation) *url.URL
}

type AuthflowV2Navigator struct {
	Endpoints   AuthflowV2NavigatorEndpointsProvider
	UIConfig    *config.UIConfig
	ErrorCookie *webapp.ErrorCookie
}

var _ handlerwebapp.AuthflowNavigator = &AuthflowV2Navigator{}

func (n *AuthflowV2Navigator) NavigateNonRecoverableError(r *http.Request, u *url.URL, e error) {
	switch {
	case user.IsAccountStatusError(e):
		u.Path = AuthflowV2RouteAccountStatus
	case errors.Is(e, api.ErrNoAuthenticator):
		u.Path = webapp.AuthflowRouteNoAuthenticator
	case errors.Is(e, authflow.ErrFlowNotFound):
		u.Path = n.Endpoints.ErrorEndpointURL(n.UIConfig.Implementation).Path
	case apierrors.IsKind(e, webapp.WebUIInvalidSession):
		// Show WebUIInvalidSession error in different page.
		u.Path = n.Endpoints.ErrorEndpointURL(n.UIConfig.Implementation).Path
	case r.Method == http.MethodGet:
		// If the request method is Get, avoid redirect back to the same path
		// which causes infinite redirect loop
		u.Path = n.Endpoints.ErrorEndpointURL(n.UIConfig.Implementation).Path
	}
}

func (n *AuthflowV2Navigator) Navigate(s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	if s.HasBranchToTake() {
		panic(fmt.Errorf("expected screen to have its branches taken"))
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

func (n *AuthflowV2Navigator) navigateSignup(s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	n.navigateSignupPromote(s, r, webSessionID, result, AuthflowV2RouteSignup)
}

func (n *AuthflowV2Navigator) navigatePromote(s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	n.navigateSignupPromote(s, r, webSessionID, result, AuthflowV2RoutePromote)
}

func (n *AuthflowV2Navigator) navigateSignupPromote(s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result, expectedPath string) {
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
			s.Advance(AuthflowV2RouteCreatePassword, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			switch data := s.StateTokenFlowResponse.Action.Data.(type) {
			case declarative.NodeVerifyClaimData:
				// 1. We do not need to enter the target.
				switch data.OTPForm {
				case otp.FormCode:
					s.Advance(AuthflowV2RouteEnterOOBOTP, result)
				case otp.FormLink:
					s.Advance(AuthflowV2RouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
				}
			case declarative.IntentSignupFlowStepCreateAuthenticatorData:
				// 2. We need to enter the target.
				s.Advance(AuthflowV2RouteSetupOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.Advance(AuthflowV2RouteSetupTOTP, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			data := s.StateTokenFlowResponse.Action.Data
			switch data := data.(type) {
			case declarative.NodeVerifyClaimData:
				// 1. We do not need to enter the target.
				channel := data.Channel
				switch channel {
				case model.AuthenticatorOOBChannelSMS:
					s.Advance(AuthflowV2RouteEnterOOBOTP, result)
				case model.AuthenticatorOOBChannelWhatsapp:
					s.Advance(AuthflowV2RouteWhatsappOTP, result)
				default:
					panic(fmt.Errorf("unexpected channel: %v", channel))
				}
			case declarative.IntentSignupFlowStepCreateAuthenticatorData:
				// 2. We need to enter the target.
				s.Advance(AuthflowV2RouteSetupOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		default:
			panic(fmt.Errorf("unexpected authentication: %v", s.StateTokenFlowResponse.Action.Authentication))
		}
	case config.AuthenticationFlowStepTypeVerify:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.NodeVerifyClaimData)
		channel := data.Channel
		switch data.OTPForm {
		case otp.FormCode:
			switch channel {
			case model.AuthenticatorOOBChannelEmail:
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelSMS:
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.Advance(AuthflowV2RouteWhatsappOTP, result)
			case "":
				// Verify may not have branches.
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case otp.FormLink:
			s.Advance(AuthflowV2RouteOOBOTPLink, result)
		}
	case config.AuthenticationFlowStepTypeFillInUserProfile:
		panic(fmt.Errorf("fill_in_user_profile is not supported yet"))
	case config.AuthenticationFlowStepTypeViewRecoveryCode:
		s.Advance(AuthflowV2RouteViewRecoveryCode, result)
	case config.AuthenticationFlowStepTypePromptCreatePasskey:
		s.Advance(AuthflowV2RoutePromptCreatePasskey, result)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowV2Navigator) navigateStepIdentify(s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result, expectedPath string) {
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
		q.Set(webapp.AuthflowQueryKey, s.Screen.StateToken.XStep)
		u.RawQuery = q.Encode()

		result.NavigationAction = "replace"
		result.RedirectURI = u.String()
	case config.AuthenticationFlowIdentificationOAuth:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.OAuthData)

		switch data.OAuthProviderType {
		case config.OAuthSSOProviderTypeWechat:
			s.Advance(AuthflowV2RouteWechat, result)
		default:
			authorizationURL, _ := url.Parse(data.OAuthAuthorizationURL)
			q := authorizationURL.Query()

			state := webapp.AuthflowOAuthState{
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

func (n *AuthflowV2Navigator) navigateLogin(s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(s, r, webSessionID, result, AuthflowV2RouteLogin)
	case config.AuthenticationFlowStepTypeAuthenticate:
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.StepAuthenticateData).Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			s.Advance(AuthflowV2RouteEnterPassword, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			switch data := s.StateTokenFlowResponse.Action.Data.(type) {
			case declarative.NodeVerifyClaimData:
				switch data.OTPForm {
				case otp.FormCode:
					s.Advance(AuthflowV2RouteEnterOOBOTP, result)
				case otp.FormLink:
					s.Advance(AuthflowV2RouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
				}
			case declarative.NodeAuthenticationOOBData:
				switch data.OTPForm {
				case otp.FormCode:
					s.Advance(AuthflowV2RouteEnterOOBOTP, result)
				case otp.FormLink:
					s.Advance(AuthflowV2RouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
				}
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.Advance(AuthflowV2RouteEnterTOTP, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			channel := s.Screen.TakenChannel
			switch channel {
			case model.AuthenticatorOOBChannelSMS:
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.Advance(AuthflowV2RouteWhatsappOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			s.Advance(AuthflowV2RouteEnterRecoveryCode, result)
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			s.Advance(AuthflowV2RouteUsePasskey, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	case config.AuthenticationFlowStepTypeCheckAccountStatus:
		s.Advance(AuthflowV2RouteAccountStatus, result)
	case config.AuthenticationFlowStepTypeTerminateOtherSessions:
		s.Advance(AuthflowV2RouteTerminateOtherSessions, result)
	case config.AuthenticationFlowStepTypeChangePassword:
		s.Advance(AuthflowV2RouteChangePassword, result)
	case config.AuthenticationFlowStepTypePromptCreatePasskey:
		s.Advance(AuthflowV2RoutePromptCreatePasskey, result)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowV2Navigator) navigateReauth(s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(s, r, webSessionID, result, AuthflowV2RouteReauth)
	case config.AuthenticationFlowStepTypeAuthenticate:
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.StepAuthenticateData).Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			s.Advance(AuthflowV2RouteEnterPassword, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			switch data := s.StateTokenFlowResponse.Action.Data.(type) {
			case declarative.NodeVerifyClaimData:
				switch data.OTPForm {
				case otp.FormCode:
					s.Advance(AuthflowV2RouteEnterOOBOTP, result)
				case otp.FormLink:
					s.Advance(AuthflowV2RouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
				}
			case declarative.NodeAuthenticationOOBData:
				switch data.OTPForm {
				case otp.FormCode:
					s.Advance(AuthflowV2RouteEnterOOBOTP, result)
				case otp.FormLink:
					s.Advance(AuthflowV2RouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
				}
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.Advance(AuthflowV2RouteEnterTOTP, result)
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			channel := s.Screen.TakenChannel
			switch channel {
			case model.AuthenticatorOOBChannelSMS:
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.Advance(AuthflowV2RouteWhatsappOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			s.Advance(AuthflowV2RouteUsePasskey, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowV2Navigator) navigateSignupLogin(s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(s, r, webSessionID, result, AuthflowV2RouteSignupLogin)
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowV2Navigator) navigateAccountRecovery(s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	navigate := func(path string, query *url.Values) {
		u := *r.URL
		u.Path = path
		q := u.Query()
		q.Set(webapp.AuthflowQueryKey, s.Screen.StateToken.XStep)
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
		navigate(AuthflowV2RouteForgotPassword, &url.Values{})
	case config.AuthenticationFlowStepTypeSelectDestination:
		navigate(AuthflowV2RouteForgotPassword, &url.Values{})
	case config.AuthenticationFlowStepTypeVerifyAccountRecoveryCode:
		data, ok := s.StateTokenFlowResponse.Action.Data.(declarative.IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData)
		if ok && data.OTPForm == declarative.AccountRecoveryOTPFormCode {
			navigate(AuthflowV2RouteForgotPasswordOTP, &url.Values{"x_can_back_to_login": []string{"true"}})
		} else {
			navigate(AuthflowV2RouteForgotPasswordSuccess, &url.Values{"x_can_back_to_login": []string{"false"}})
		}
	case config.AuthenticationFlowStepTypeResetPassword:
		navigate(AuthflowV2RouteResetPassword, &url.Values{})
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}
