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
	"error-password-reset-failed": "This reset password link is invalid, used or expired. Please request a new one.",

	"back-button-title": "Back",
	"next-button-label": "Next",
	"connect-button-label": "Connect",
	"disconnect-button-label": "Disconnect",
	"change-button-label": "Change",

	"sign-in-apple": "Sign in with Apple",
	"sign-up-apple": "Sign up with Apple",
	"sign-in-google": "Sign in with Google",
	"sign-up-google": "Sign up with Google",
	"sign-in-facebook": "Sign in with Facebook",
	"sign-up-facebook": "Sign up with Facebook",
	"sign-in-linkedin": "Sign in with LinkedIn",
	"sign-up-linkedin": "Sign up with LinkedIn",
	"sign-in-azureadv2": "Sign in with Azure AD",
	"sign-up-azureadv2": "Sign up with Azure AD",
	"sso-login-id-separator": "or",

	"phone-number-placeholder": "Phone number",
	"login-id-placeholder": "Email or username",
	"use-text-login-id-description": "Use email or username instead",
	"use-email-login-id-description": "Use email instead",
	"use-phone-login-id-description": "Use phone number instead",
	"signup-button-hint": "Don''t have an account yet? ",
	"signup-button-label": "Create one!",
	"forgot-password-button-label": "Can''t access your account?",

	"enter-password-page-title": "Enter password",
	"password-placeholder": "Password",
	"forgot-password-button-label--enter-password-page": "Forgot Password?",
	"show-password": "Show Password",
	"hide-password": "Hide Password",

	"oob-otp-page-title--sms": "SMS Verification",
	"oob-otp-page-title--email": "Email Verification",
	"oob-otp-placeholder": "Enter code",
	"oob-otp-description--sms": "We have sent a {0} digit code to +{1}{2}. Please enter the code below to continue",
	"oob-otp-description--email": "We have sent a {0} digit code to {1}. Please enter the code below to continue",
	"oob-otp-resend-button-hint": "Didn''t receive the code? ",
	"oob-otp-resend-button-label": "Resend (0s)",
	
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

	"forgot-password-page-title": "Forgot Password",
	"email-placeholder": "Email",
	"forgot-password-email-description": "Enter your email to request instruction to reset your password",
	"forgot-password-phone-description": "Enter your phone to request instruction to reset your password",
	"forgot-password-success-page-title": "Request received",
	"forgot-password-success-description": "If you have an account, please follow the instruction sent to {0} to reset your password",
	"login-button-label--forgot-password-success-page": "Sign in",

	"reset-password-page-title": "Reset Password",
	"reset-password-description": "Please enter your new password below.",

	"reset-password-success-page-title": "Password Reset",
	"reset-password-success-description": "You have successfully reset your password. You can now sign in with it.",

	"logout-button-hint": "To logout, please click the button below.",
	"logout-button-label": "Logout",

	"settings-identity-title": "Account settings",
	"settings-identity-oauth-google": "Google",
	"settings-identity-oauth-apple": "Apple",
	"settings-identity-oauth-facebook": "Facebook",
	"settings-identity-oauth-linkedin": "LinkedIn",
	"settings-identity-oauth-azureadv2": "Azure AD",
	"settings-identity-login-id-email": "Email Address",
	"settings-identity-login-id-phone": "Phone Number",
	"settings-identity-login-id-username": "Username",
	"settings-identity-login-id-raw": "Username",

	"enter-login-id-page-title--change": "Change your {0}",
	"enter-login-id-page-title--add": "Enter your {0}"
	}`,
}
