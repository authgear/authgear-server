package webapp

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeAuthUITranslationJSON config.TemplateItemType = "auth_ui_translation.json"
)

var TemplateAuthUITranslationJSON = template.Spec{
	Type: TemplateItemTypeAuthUITranslationJSON,
	Default: `{
	"error-email-or-username-required": "Email or Username is required",
	"error-password-required": "Password is required",
	"error-calling-code-required": "Calling code is required",
	"error-phone-number-required": "Phone number is required",
	"error-phone-number-format": "Phone number must contains digits only",
	"error-invalid-email": "Invalid email address",
	"error-invalid-username": "Invalid username",
	"error-invalid-credentials": "Incorrect email, phone number, username, or password",

	"back-button-title": "Back",

	"sign-in-apple": "Sign in with Apple",
	"sign-in-google": "Sign in with Google",
	"sign-in-facebook": "Sign in with Facebook",
	"sign-in-instagram": "Sign in with Instagram",
	"sign-in-linkedin": "Sign in with LinkedIn",
	"sign-in-azureadv2": "Sign in with Azure AD",
	"sso-login-id-separator": "or",

	"phone-number-placeholder": "Phone number",
	"login-id-placeholder": "Email or username",
	"use-text-login-id-description": "Use email or username instead",
	"use-phone-login-id-description": "Use phone number instead",
	"signup-button-hint": "Don''t have an account yet? ",
	"signup-button-label": "Create one!",
	"forgot-password-button-label": "Can''t access your account?",
	"confirm-login-id-button-label": "Next",

	"enter-password-page-title": "Enter password",
	"password-placeholder": "Password",
	"forgot-password-button-label--enter-password-page": "Forgot Password?",
	"confirm-password-button-label": "Next",
	"show-password": "Show Password",
	"hide-password": "Hide Password",
	
	"use-login-id-key": "Use {0} instead",
	"login-button-hint": "Have an account already? ",
	"login-button-label": "Sign in!",

	"create-password-page-title": "Create password",
	"password-policy-minimum-length": "At least {0, plural, one{# character} other{# characters}} long",
	"password-policy-uppercase": "At least 1 uppercase character",
	"password-policy-lowercase": "At least 1 lowercase character",
	"password-policy-digit": "At least 1 digit",
	"password-policy-symbol": "At least 1 symbol",
	"password-policy-banned-words": "NO banned words",
	"password-policy-guessable-level-1": "NOT too guessable",
	"password-policy-guessable-level-2": "NOT very guessable",
	"password-policy-guessable-level-3": "NOT somewhat guessable",
	"password-policy-guessable-level-4": "Safely unguessable",
	"password-policy-guessable-level-5": "Very unguessable",
	"sms-charge-warning": "By providing your phone number, you agree to receive service notifications to your mobile phone. Text messaging rates may apply.",

	"logout-button-hint": "To logout, please click the button below.",
	"logout-button-label": "Logout"
	}`,
}
