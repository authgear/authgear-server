package authflowv2

import (
	"context"
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
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
	"github.com/authgear/authgear-server/pkg/lib/webappoauth"
)

const (
	AuthflowV2RouteLogin               = "/login"
	AuthflowV2RouteSignup              = "/signup"
	AuthflowV2RoutePromote             = "/flows/promote_user"
	AuthflowV2RouteReauth              = "/reauth"
	AuthflowV2RouteSelectAccount       = "/authflow/v2/select_account"
	AuthflowV2RouteVerifyBotProtection = "/authflow/v2/verify_bot_protection"
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
	AuthflowV2RouteChangePasswordSuccess = "/authflow/v2/change_password/success"
	// nolint: gosec
	AuthflowV2RouteEnterPassword     = "/authflow/v2/enter_password"
	AuthflowV2RouteEnterRecoveryCode = "/authflow/v2/enter_recovery_code"
	AuthflowV2RouteEnterOOBOTP       = "/authflow/v2/enter_oob_otp"
	AuthflowV2RouteOOBOTPLink        = "/authflow/v2/oob_otp_link"
	AuthflowV2RouteVerifyLink        = "/authflow/v2/verify_login_link"
	AuthflowV2RouteEnterTOTP         = "/authflow/v2/enter_totp"
	AuthflowV2RouteSetupTOTP         = "/authflow/v2/setup_totp"
	AuthflowV2RouteSetupOOBOTP       = "/authflow/v2/setup_oob_otp"

	AuthflowV2RouteLDAPLogin = "/authflow/v2/ldap_login"
	// nolint: gosec
	AuthflowV2RouteUsePasskey = "/authflow/v2/use_passkey"
	// nolint: gosec
	AuthflowV2RouteForgotPassword = "/authflow/v2/forgot_password"
	// nolint: gosec
	AuthflowV2RouteForgotPasswordOTP = "/authflow/v2/forgot_password/otp"
	// nolint: gosec
	AuthflowV2RouteForgotPasswordLinkSent = "/authflow/v2/forgot_password/link/sent"
	// nolint: gosec
	AuthflowV2RouteResetPassword = "/authflow/v2/reset_password"
	// nolint: gosec
	AuthflowV2RouteResetPasswordSuccess = "/authflow/v2/reset_password/success"
	AuthflowV2RouteWechat               = "/authflow/v2/wechat"
	AuthflowV2RouteAccountLinking       = "/authflow/v2/account_linking"
	// nolint:gosec
	AuthflowV2RouteOAuthProviderDemoCredential = "/authflow/v2/oauth_provider_demo_credential"

	SettingsV2RouteSettings                   = "/settings"
	SettingsV2RouteProfilePictureEditSettings = "/settings/profile/picture/edit"

	SettingsV2RouteSettingsProfileGenderEdit = "/settings/profile/gender/edit"

	SettingsV2RouteAdvancedSettings = "/settings/advanced_settings"

	// The following routes are dead ends.
	AuthflowV2RouteAccountStatus   = "/authflow/v2/account_status"
	AuthflowV2RouteNoAuthenticator = "/authflow/v2/no_authenticator"
	// nolint:gosec
	AuthflowV2RouteOAuthProviderMissingCredentials = "/authflow/v2/oauth_provider_missing_credential"

	AuthflowV2RouteFinishFlow = "/authflow/v2/finish"

	AuthflowV2RouteSettingsProfile             = "/settings/v2/profile"
	AuthflowV2RouteSettingsMFA                 = "/settings/mfa"
	AuthflowV2RouteSettingsMFAViewRecoveryCode = "/settings/mfa/view_recovery_code"

	// nolint: gosec
	AuthflowV2RouteSettingsMFACreatePassword = "/settings/mfa/create_password"
	AuthflowV2RouteSettingsMFACreateOOBOTP   = "/settings/mfa/create_oob_otp_:channel"
	AuthflowV2RouteSettingsMFAEnterOOBOTP    = "/settings/mfa/enter_oob_otp"
	// nolint: gosec
	AuthflowV2RouteSettingsMFAPassword = "/settings/mfa/password"
	// nolint: gosec
	AuthflowV2RouteSettingsMFAChangePassword = "/settings/mfa/change_password"
	AuthflowV2RouteSettingsMFACreateTOTP     = "/settings/mfa/create_totp"
	AuthflowV2RouteSettingsMFAEnterTOTP      = "/settings/mfa/enter_totp"

	AuthflowV2RouteSettingsIdentityListEmail          = "/settings/identity/email"
	AuthflowV2RouteSettingsIdentityAddEmail           = "/settings/identity/add_email"
	AuthflowV2RouteSettingsIdentityChangeEmail        = "/settings/identity/change_email"
	AuthflowV2RouteSettingsIdentityViewEmail          = "/settings/identity/view_email"
	AuthflowV2RouteSettingsIdentityChangePrimaryEmail = "/settings/identity/change_primary_email"
	AuthflowV2RouteSettingsIdentityVerifyEmail        = "/settings/identity/verify_email"

	AuthflowV2RouteSettingsIdentityListPhone          = "/settings/identity/phone"
	AuthflowV2RouteSettingsIdentityAddPhone           = "/settings/identity/add_phone"
	AuthflowV2RouteSettingsIdentityChangePhone        = "/settings/identity/change_phone"
	AuthflowV2RouteSettingsIdentityViewPhone          = "/settings/identity/view_phone"
	AuthflowV2RouteSettingsIdentityChangePrimaryPhone = "/settings/identity/change_primary_phone"
	AuthflowV2RouteSettingsIdentityVerifyPhone        = "/settings/identity/verify_phone"

	AuthflowV2RouteSettingsIdentityListUsername   = "/settings/identity/username"
	AuthflowV2RouteSettingsIdentityAddUsername    = "/settings/identity/add_username"
	AuthflowV2RouteSettingsIdentityViewUsername   = "/settings/identity/view_username"
	AuthflowV2RouteSettingsIdentityChangeUsername = "/settings/identity/change_username"

	AuthflowV2RouteSettingsIdentityListOAuth     = "/settings/identity/oauth"
	AuthflowV2RouteSettingsIdentityOAuthCallback = "/settings/identity/oauth/callback"
)

type AuthflowV2NavigatorEndpointsProvider interface {
	ErrorEndpointURL() *url.URL
	SelectAccountEndpointURL() *url.URL
	VerifyBotProtectionEndpointURL() *url.URL
}

type AuthflowV2NavigatorOAuthStateStore interface {
	GenerateState(ctx context.Context, state *webappoauth.WebappOAuthState) (stateToken string, err error)
}

type AuthflowV2Navigator struct {
	AppID           config.AppID
	Endpoints       AuthflowV2NavigatorEndpointsProvider
	OAuthStateStore AuthflowV2NavigatorOAuthStateStore
}

var _ handlerwebapp.AuthflowNavigator = &AuthflowV2Navigator{}

func (n *AuthflowV2Navigator) NavigateNonRecoverableError(r *http.Request, u *url.URL, e error) {
	switch {
	case user.IsAccountStatusError(e):
		u.Path = AuthflowV2RouteAccountStatus
	case errors.Is(e, api.ErrNoAuthenticator):
		u.Path = AuthflowV2RouteNoAuthenticator
	case errors.Is(e, authflow.ErrFlowNotFound):
		u.Path = n.Endpoints.ErrorEndpointURL().Path
	case apierrors.IsKind(e, api.OAuthProviderMissingCredentials):
		u.Path = AuthflowV2RouteOAuthProviderMissingCredentials
	case apierrors.IsKind(e, webapp.WebUIInvalidSession):
		// Show WebUIInvalidSession error in different page.
		u.Path = n.Endpoints.ErrorEndpointURL().Path
	case r.Method == http.MethodGet:
		// If the request method is Get, avoid redirect back to the same path
		// which causes infinite redirect loop
		u.Path = n.Endpoints.ErrorEndpointURL().Path
	}
}

func (n *AuthflowV2Navigator) NavigateResetPasswordSuccessPage() string {
	return AuthflowV2RouteResetPasswordSuccess
}

func (n *AuthflowV2Navigator) navigateWithScreen(path string, screen *webapp.AuthflowScreen, r *http.Request, query *url.Values) (result *webapp.Result) {
	u := *r.URL
	u.Path = path
	q := u.Query()
	q.Set(webapp.AuthflowQueryKey, screen.StateToken.XStep)
	for k, param := range *query {
		for _, p := range param {
			q.Add(k, p)
		}
	}
	u.RawQuery = q.Encode()
	result = &webapp.Result{}
	result.NavigationAction = webapp.NavigationActionAdvance
	result.RedirectURI = u.String()
	return result
}

func (n *AuthflowV2Navigator) NavigateChangePasswordSuccessPage(s *webapp.AuthflowScreen, r *http.Request, webSessionID string) (result *webapp.Result) {
	return n.navigateWithScreen(AuthflowV2RouteChangePasswordSuccess, s, r, &url.Values{})
}

func (n *AuthflowV2Navigator) NavigateOAuthProviderDemoCredentialPage(s *webapp.AuthflowScreen, r *http.Request) (result *webapp.Result) {
	return n.navigateWithScreen(AuthflowV2RouteOAuthProviderDemoCredential, s, r, &url.Values{})
}

func (n *AuthflowV2Navigator) Navigate(ctx context.Context, s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	if s.HasBranchToTake() {
		panic(fmt.Errorf("expected screen to have its branches taken"))
	}

	if s.StateTokenFlowResponse.Action.Type == authflow.FlowActionTypeFinished {
		s.RedirectToFinish(AuthflowV2RouteFinishFlow, result)
		return
	}

	switch s.StateTokenFlowResponse.Type {
	case authflow.FlowTypeSignup:
		n.navigateSignup(ctx, s, r, webSessionID, result)
	case authflow.FlowTypePromote:
		n.navigatePromote(ctx, s, r, webSessionID, result)
	case authflow.FlowTypeLogin:
		n.navigateLogin(ctx, s, r, webSessionID, result)
	case authflow.FlowTypeSignupLogin:
		n.navigateSignupLogin(ctx, s, r, webSessionID, result)
	case authflow.FlowTypeReauth:
		n.navigateReauth(ctx, s, r, webSessionID, result)
	case authflow.FlowTypeAccountRecovery:
		n.navigateAccountRecovery(s, r, webSessionID, result)
	default:
		panic(fmt.Errorf("unexpected flow type: %v", s.StateTokenFlowResponse.Type))
	}
}

func (n *AuthflowV2Navigator) navigateSignup(ctx context.Context, s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	n.navigateSignupPromote(ctx, s, r, webSessionID, result, AuthflowV2RouteSignup)
}

func (n *AuthflowV2Navigator) navigatePromote(ctx context.Context, s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	n.navigateSignupPromote(ctx, s, r, webSessionID, result, AuthflowV2RoutePromote)
}

func (n *AuthflowV2Navigator) navigateSignupPromote(ctx context.Context, s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result, expectedPath string) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(ctx, s, r, webSessionID, result, expectedPath)
	case config.AuthenticationFlowStepTypeCreateAuthenticator:
		authentication := getTakenBranchCreateAuthenticatorAuthentication(s)
		switch authentication {
		case model.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryPassword:
			s.Advance(AuthflowV2RouteCreatePassword, result)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			switch data := s.StateTokenFlowResponse.Action.Data.(type) {
			case declarative.VerifyOOBOTPData:
				// 1. We do not need to enter the target.
				switch data.OTPForm {
				case otp.FormCode:
					s.Advance(AuthflowV2RouteEnterOOBOTP, result)
				case otp.FormLink:
					s.Advance(AuthflowV2RouteOOBOTPLink, result)
				default:
					panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
				}
			case declarative.CreateAuthenticatorData:
				// 2. We need to enter the target.
				s.Advance(AuthflowV2RouteSetupOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected data: %T", s.StateTokenFlowResponse.Action.Data))
			}
		case model.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.Advance(AuthflowV2RouteSetupTOTP, result)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			data := s.StateTokenFlowResponse.Action.Data
			switch data := data.(type) {
			case declarative.VerifyOOBOTPData:
				// 1. We do not need to enter the target.
				channel := data.Channel
				switch channel {
				case model.AuthenticatorOOBChannelSMS:
					s.Advance(AuthflowV2RouteEnterOOBOTP, result)
				case model.AuthenticatorOOBChannelWhatsapp:
					s.Advance(AuthflowV2RouteEnterOOBOTP, result)
				default:
					panic(fmt.Errorf("unexpected channel: %v", channel))
				}
			case declarative.CreateAuthenticatorData:
				// 2. We need to enter the target.
				s.Advance(AuthflowV2RouteSetupOOBOTP, result)
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
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelSMS:
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
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

func (n *AuthflowV2Navigator) navigateStepIdentify(ctx context.Context, s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result, expectedPath string) {
	if _, ok := s.StateTokenFlowResponse.Action.Data.(declarative.AccountLinkingIdentifyData); ok {
		s.Advance(AuthflowV2RouteAccountLinking, result)
		return
	}

	identification := s.StateTokenFlowResponse.Action.Identification
	switch identification {
	case "":
		fallthrough
	case model.AuthenticationFlowIdentificationIDToken:
		fallthrough
	case model.AuthenticationFlowIdentificationEmail:
		fallthrough
	case model.AuthenticationFlowIdentificationPhone:
		fallthrough
	case model.AuthenticationFlowIdentificationUsername:
		fallthrough
	case model.AuthenticationFlowIdentificationPasskey:
		// Redirect to the expected path with x_step set.
		u := *r.URL
		// The current authflow state is 1 to 1 mapping with the path.
		// Since Both LDAP_login & login page share the same state
		// We workaround for this exceptional case.
		if u.Path != AuthflowV2RouteLDAPLogin {
			u.Path = expectedPath
		}
		q := u.Query()
		q.Set(webapp.AuthflowQueryKey, s.Screen.StateToken.XStep)
		u.RawQuery = q.Encode()

		result.NavigationAction = webapp.NavigationActionReplace
		result.RedirectURI = u.String()
	case model.AuthenticationFlowIdentificationOAuth:
		data := s.StateTokenFlowResponse.Action.Data.(declarative.OAuthData)

		switch data.OAuthProviderType {
		case wechat.Type:
			s.Advance(AuthflowV2RouteWechat, result)
		default:
			authorizationURL, _ := url.Parse(data.OAuthAuthorizationURL)
			q := authorizationURL.Query()
			// Back to the current screen if error
			errorRedirectURI := url.URL{Path: r.URL.Path, RawQuery: r.URL.Query().Encode()}

			state := &webappoauth.WebappOAuthState{
				AppID:            string(n.AppID),
				WebSessionID:     webSessionID,
				UIImplementation: config.UIImplementationAuthflowV2,
				XStep:            s.Screen.StateToken.XStep,
				ErrorRedirectURI: errorRedirectURI.String(),
				ProviderAlias:    data.Alias,
			}
			stateToken, err := n.OAuthStateStore.GenerateState(ctx, state)
			if err != nil {
				panic(err)
			}

			q.Set("state", stateToken)
			authorizationURL.RawQuery = q.Encode()

			result.NavigationAction = webapp.NavigationActionRedirect
			result.RedirectURI = authorizationURL.String()
		}
	case model.AuthenticationFlowIdentificationLDAP:
		// Not expected to trigger this case
		panic(fmt.Errorf("not expected to trigger: %v", identification))
	default:
		panic(fmt.Errorf("unexpected identification: %v", identification))
	}
}

func (n *AuthflowV2Navigator) navigateLoginStepAuthenticate(s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	switch s.StateTokenFlowResponse.Action.Data.(type) {
	case declarative.IntentCreateAuthenticatorTOTPData:
		s.Advance(AuthflowV2RouteSetupTOTP, result)
	case declarative.CreateAuthenticatorData:
		authentication := getTakenBranchCreateAuthenticatorAuthentication(s)
		switch authentication {
		case model.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryPassword:
			s.Advance(AuthflowV2RouteCreatePassword, result)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			s.Advance(AuthflowV2RouteSetupOOBOTP, result)
		case model.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.Advance(AuthflowV2RouteSetupTOTP, result)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			s.Advance(AuthflowV2RouteSetupOOBOTP, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", s.StateTokenFlowResponse.Action.Authentication))
		}
	case declarative.StepAuthenticateData:
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.StepAuthenticateData).Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case model.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryPassword:
			s.Advance(AuthflowV2RouteEnterPassword, result)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			// Action data type should be VerifyOOBOTPData
			panic(fmt.Errorf("unexpected data type: %T", s.StateTokenFlowResponse.Action.Data))
		case model.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.Advance(AuthflowV2RouteEnterTOTP, result)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			channel := s.Screen.TakenChannel
			switch channel {
			case model.AuthenticatorOOBChannelSMS:
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case model.AuthenticationFlowAuthenticationRecoveryCode:
			s.Advance(AuthflowV2RouteEnterRecoveryCode, result)
		case model.AuthenticationFlowAuthenticationPrimaryPasskey:
			s.Advance(AuthflowV2RouteUsePasskey, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	// Below code is only reachable if the step requires captcha, since VerifyBotProtection screen did not use TakeBranchResultInput to feed input
	case declarative.VerifyOOBOTPData:
		var data = s.StateTokenFlowResponse.Action.Data.(declarative.VerifyOOBOTPData)
		switch data.OTPForm {
		case otp.FormCode:
			s.Advance(AuthflowV2RouteEnterOOBOTP, result)
		case otp.FormLink:
			s.Advance(AuthflowV2RouteOOBOTPLink, result)
		default:
			panic(fmt.Errorf("unexpected otp form: %v", data.OTPForm))
		}
	default:
		panic(fmt.Errorf("unexpected data type: %T", s.StateTokenFlowResponse.Action.Data))
	}
}

func (n *AuthflowV2Navigator) navigateLogin(ctx context.Context, s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	if s.Screen.IsBotProtectionRequired {
		s.Advance(AuthflowV2RouteVerifyBotProtection, result)
		return
	}
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(ctx, s, r, webSessionID, result, AuthflowV2RouteLogin)
	case config.AuthenticationFlowStepTypeAuthenticate:
		n.navigateLoginStepAuthenticate(s, r, webSessionID, result)
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

func (n *AuthflowV2Navigator) navigateReauth(ctx context.Context, s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	if s.Screen.IsBotProtectionRequired {
		s.Advance(AuthflowV2RouteVerifyBotProtection, result)
		return
	}
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(ctx, s, r, webSessionID, result, AuthflowV2RouteReauth)
	case config.AuthenticationFlowStepTypeAuthenticate:
		options := s.BranchStateTokenFlowResponse.Action.Data.(declarative.StepAuthenticateData).Options
		index := *s.Screen.TakenBranchIndex
		option := options[index]
		switch option.Authentication {
		case model.AuthenticationFlowAuthenticationPrimaryPassword:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryPassword:
			s.Advance(AuthflowV2RouteEnterPassword, result)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			switch data := s.StateTokenFlowResponse.Action.Data.(type) {
			case declarative.VerifyOOBOTPData:
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
		case model.AuthenticationFlowAuthenticationSecondaryTOTP:
			s.Advance(AuthflowV2RouteEnterTOTP, result)
		case model.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			fallthrough
		case model.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			channel := s.Screen.TakenChannel
			switch channel {
			case model.AuthenticatorOOBChannelSMS:
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			case model.AuthenticatorOOBChannelWhatsapp:
				s.Advance(AuthflowV2RouteEnterOOBOTP, result)
			default:
				panic(fmt.Errorf("unexpected channel: %v", channel))
			}
		case model.AuthenticationFlowAuthenticationPrimaryPasskey:
			s.Advance(AuthflowV2RouteUsePasskey, result)
		default:
			panic(fmt.Errorf("unexpected authentication: %v", option.Authentication))
		}
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowV2Navigator) navigateSignupLogin(ctx context.Context, s *webapp.AuthflowScreenWithFlowResponse, r *http.Request, webSessionID string, result *webapp.Result) {
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		n.navigateStepIdentify(ctx, s, r, webSessionID, result, AuthflowV2RouteSignupLogin)
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
		result.NavigationAction = webapp.NavigationActionReplace
		result.RedirectURI = u.String()
	}
	switch config.AuthenticationFlowStepType(s.StateTokenFlowResponse.Action.Type) {
	case config.AuthenticationFlowStepTypeIdentify:
		navigate(AuthflowV2RouteForgotPassword, &url.Values{})
	case config.AuthenticationFlowStepTypeSelectDestination:
		navigate(AuthflowV2RouteForgotPassword, &url.Values{})
	case config.AuthenticationFlowStepTypeVerifyAccountRecoveryCode:
		data, ok := s.StateTokenFlowResponse.Action.Data.(declarative.IntentAccountRecoveryFlowStepVerifyAccountRecoveryCodeData)
		if !ok {
			panic(fmt.Errorf("unexpected data type in step verify_account_recovery_code"))
		}
		switch data.OTPForm {
		case declarative.AccountRecoveryOTPFormCode:
			navigate(AuthflowV2RouteForgotPasswordOTP, &url.Values{"x_can_back_to_login": []string{"true"}})
		case declarative.AccountRecoveryOTPFormLink:
			navigate(AuthflowV2RouteForgotPasswordLinkSent, &url.Values{"x_can_back_to_login": []string{"false"}})
		default:
			panic(fmt.Errorf("unexpected otp form in step verify_account_recovery_code"))
		}
	case config.AuthenticationFlowStepTypeResetPassword:
		navigate(AuthflowV2RouteResetPassword, &url.Values{})
	default:
		panic(fmt.Errorf("unexpected action type: %v", s.StateTokenFlowResponse.Action.Type))
	}
}

func (n *AuthflowV2Navigator) NavigateSelectAccount(result *webapp.Result) {
	url := n.Endpoints.SelectAccountEndpointURL()
	result.RedirectURI = url.String()
}

func (n *AuthflowV2Navigator) NavigateVerifyBotProtection(result *webapp.Result) {
	url := n.Endpoints.VerifyBotProtectionEndpointURL()
	result.RedirectURI = url.String()
}
