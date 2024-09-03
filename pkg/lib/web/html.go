package web

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

func init() {
	ComponentsHTML = append(ComponentsHTML, BaseComponentsHTML...)
	ComponentsHTML = append(ComponentsHTML, TemplateWebAuthflowV2PageFrameHTML)
	SettingsComponentsHTML = append(SettingsComponentsHTML, BaseComponentsHTML...)
	SettingsComponentsHTML = append(SettingsComponentsHTML, TemplateWebSettingsV2PageFrameHTML)
}

var TemplateWebHTMLHeadHTML = template.RegisterHTML("web/__html_head.html")
var TemplateWebGeneratedAssetHTML = template.RegisterHTML("web/__generated_asset.html")
var TemplateWebHeaderHTML = template.RegisterHTML("web/__header.html")
var TemplateWebNavBarHTML = template.RegisterHTML("web/__nav_bar.html")
var TemplateWebErrorHTML = template.RegisterHTML("web/__error.html")
var TemplateWebMessageBarHTML = template.RegisterHTML("web/__message_bar.html")
var TemplateWebAlternativeStepsHTML = template.RegisterHTML("web/__alternatives.html")
var TemplateWebPhoneOTPAlternativeStepsHTML = template.RegisterHTML("web/__phone_otp_alternatives.html")
var TemplateWebUseRecoveryCodeHTML = template.RegisterHTML("web/__use_recovery_code.html")
var TemplateWebPasswordPolicyHTML = template.RegisterHTML("web/__password_policy.html")
var TemplateWebPageFrameHTML = template.RegisterHTML("web/__page_frame.html")
var TemplateWebWidePageFrameHTML = template.RegisterHTML("web/__wide_page_frame.html")
var TemplateWebModalHTML = template.RegisterHTML("web/__modal.html")
var TemplateWebWatermarkHTML = template.RegisterHTML("web/__watermark.html")
var TemplateWebRecoveryCodeHTML = template.RegisterHTML("web/__recovery_code.html")
var TemplateWebPasswordInputHTML = template.RegisterHTML("web/__password_input.html")
var TemplateWebPasswordStrengthMeterHTML = template.RegisterHTML("web/__password_strength_meter.html")
var TemplateWebTutorialHTML = template.RegisterHTML("web/__tutorial.html")
var TemplateWebTermsOfServiceAndPrivacyPolicyFooterHTML = template.RegisterHTML("web/__toc_pp_footer.html")
var TemplateWebAuthflowBranchHTML = template.RegisterHTML("web/__authflow_branch.html")
var TemplateWebAuthflowForgotPasswordAlternativesHTML = template.RegisterHTML("web/__authflow_forgot_password_alternatives.html")
var TemplateWebTranslationMessageHTML = template.RegisterHTML("web/__translation_message.html")

// TODO: This file could be overridable per app, depends on the project plan in future
var TemplateWebAuthflowV2LayoutHTML = template.RegisterHTML("web/authflowv2/layout.html")
var TemplateWebAuthflowV2SettingsLayoutHTML = template.RegisterHTML("web/authflowv2/settings_layout.html")

var TemplateWebAuthflowV2HTMLHeadHTML = template.RegisterHTML("web/authflowv2/__html_head.html")
var TemplateWebAuthflowV2LoadBotProtectionHTML = template.RegisterHTML("web/authflowv2/__load_bot_protection.html")
var TemplateWebAuthflowV2GeneratedAssetHTML = template.RegisterHTML("web/authflowv2/__generated_asset.html")
var TemplateWebAuthflowV2BasePageFrameHTML = template.RegisterHTML("web/authflowv2/__base_page_frame.html")
var TemplateWebAuthflowV2PageFrameHTML = template.RegisterHTML("web/authflowv2/__page_frame.html")
var TemplateWebAuthflowV2DialogHTML = template.RegisterHTML("web/authflowv2/__dialog.html")
var TemplateWebAuthflowV2BotProtectionWidgetHTML = template.RegisterHTML("web/authflowv2/__bot_protection_widget.html")
var TemplateWebAuthflowV2BotProtectionFormInputHTML = template.RegisterHTML("web/authflowv2/__bot_protection_form_input.html")
var TemplateWebAuthflowV2BotProtectionControllerHTML = template.RegisterHTML("web/authflowv2/__bot_protection_controller.html")
var TemplateWebAuthflowV2BotProtectionControllerAttrHTML = template.RegisterHTML("web/authflowv2/__bot_protection_controller_attr.html")
var TemplateWebAuthflowV2BotProtectionDialogHTML = template.RegisterHTML("web/authflowv2/__bot_protection_dialog.html")
var TemplateWebAuthflowV2BrandLogoHTML = template.RegisterHTML("web/authflowv2/__brand_logo.html")
var TemplateWebAuthflowV2HeaderHTML = template.RegisterHTML("web/authflowv2/__header.html")
var TemplateWebAuthflowV2DividerHTML = template.RegisterHTML("web/authflowv2/__divider.html")
var TemplateWebAuthflowV2AlertMessageHTML = template.RegisterHTML("web/authflowv2/__alert_message.html")
var TemplateWebAuthflowV2OTPInputHTML = template.RegisterHTML("web/authflowv2/__otp_input.html")
var TemplateWebAuthflowV2PasswordInputHTML = template.RegisterHTML("web/authflowv2/__password_input.html")
var TemplateWebAuthflowV2PasswordFieldHTML = template.RegisterHTML("web/authflowv2/__password_field.html")
var TemplateWebAuthflowV2NewPasswordFieldHTML = template.RegisterHTML("web/authflowv2/__new_password_field.html")
var TemplateWebAuthflowV2PasswordStrengthMeterHTML = template.RegisterHTML("web/authflowv2/__password_strength_meter.html")
var TemplateWebAuthflowV2PhoneInputHTML = template.RegisterHTML("web/authflowv2/__phone_input.html")
var TemplateWebAuthflowV2ErrorHTML = template.RegisterHTML("web/authflowv2/__error.html")
var TemplateWebAuthflowV2PasswordPolicyHTML = template.RegisterHTML("web/authflowv2/__password_policy.html")
var TemplateWebAuthflowV2BranchHTML = template.RegisterHTML("web/authflowv2/__authflow_branch.html")
var TemplateWebAuthflowV2LockoutHTML = template.RegisterHTML("web/authflowv2/__lockout.html")
var TemplateWebAuthflowV2ForgotPasswordAlternativesHTML = template.RegisterHTML("web/authflowv2/__forgot_password_alternatives.html")
var TemplateWebAuthflowV2ErrorPageLayoutHTML = template.RegisterHTML("web/authflowv2/__error_page_layout.html")
var TemplateWebAuthflowV2DeviceTokenCheckboxHTML = template.RegisterHTML("web/authflowv2/__device_token_checkbox.html")
var TemplateWebAuthflowV2TermsOfServiceAndPrivacyPolicyFooterHTML = template.RegisterHTML("web/authflowv2/__toc_pp_footer.html")
var TemplateWebAuthflowV2WatermarkHTML = template.RegisterHTML("web/authflowv2/__watermark.html")
var TemplateWebAuthflowV2CSRFErrorPageLayoutHTML = template.RegisterHTML("web/authflowv2/__csrf_error_page_layout.html")
var TemplateWebAuthflowV2SettingsHeaderHTML = template.RegisterHTML("web/authflowv2/__settings_header.html")
var TemplateWebAuthflowV2SettingsItemHTML = template.RegisterHTML("web/authflowv2/__settings_item.html")
var TemplateWebAuthflowV2SettingsRadioHTML = template.RegisterHTML("web/authflowv2/__settings_radio.html")

var TemplateWebSettingsV2PageFrameHTML = template.RegisterHTML("web/authflowv2/__settings_page_frame.html")
var TemplateWebAuthflowV2NavBar = template.RegisterHTML("web/authflowv2/__navbar.html")
var TemplateWebAuthflowV2DateInputHTML = template.RegisterHTML("web/authflowv2/__date_input.html")
var TemplateWebAuthflowV2UserProfilePic = template.RegisterHTML("web/authflowv2/__settings_user_profile_pic.html")
var TemplateWebAuthflowV2SettingsTextInput = template.RegisterHTML("web/authflowv2/__settings_text_input.html")
var TemplateWebAuthflowV2LocaleInputHTML = template.RegisterHTML("web/authflowv2/__locale_input.html")
var TemplateWebAuthflowV2TimezoneInput = template.RegisterHTML("web/authflowv2/__timezone_input.html")
var TemplateWebAuthflowV2CountryInput = template.RegisterHTML("web/authflowv2/__country_input.html")

var BaseComponentsHTML = []*template.HTML{
	TemplateWebHTMLHeadHTML,
	TemplateWebGeneratedAssetHTML,
	TemplateWebHeaderHTML,
	TemplateWebNavBarHTML,
	TemplateWebErrorHTML,
	TemplateWebMessageBarHTML,
	TemplateWebAlternativeStepsHTML,
	TemplateWebPhoneOTPAlternativeStepsHTML,
	TemplateWebUseRecoveryCodeHTML,
	TemplateWebPasswordPolicyHTML,
	TemplateWebPageFrameHTML,
	TemplateWebWidePageFrameHTML,
	TemplateWebModalHTML,
	TemplateWebWatermarkHTML,
	TemplateWebRecoveryCodeHTML,
	TemplateWebPasswordInputHTML,
	TemplateWebPasswordStrengthMeterHTML,
	TemplateWebTutorialHTML,
	TemplateWebTermsOfServiceAndPrivacyPolicyFooterHTML,
	TemplateWebAuthflowBranchHTML,
	TemplateWebAuthflowForgotPasswordAlternativesHTML,
	TemplateWebTranslationMessageHTML,

	TemplateWebAuthflowV2LayoutHTML,
	TemplateWebAuthflowV2SettingsLayoutHTML,
	TemplateWebAuthflowV2HTMLHeadHTML,
	TemplateWebAuthflowV2LoadBotProtectionHTML,
	TemplateWebAuthflowV2GeneratedAssetHTML,
	TemplateWebAuthflowV2BasePageFrameHTML,
	TemplateWebAuthflowV2DialogHTML,
	TemplateWebAuthflowV2BotProtectionWidgetHTML,
	TemplateWebAuthflowV2BotProtectionFormInputHTML,
	TemplateWebAuthflowV2BotProtectionControllerHTML,
	TemplateWebAuthflowV2BotProtectionControllerAttrHTML,
	TemplateWebAuthflowV2BotProtectionDialogHTML,
	TemplateWebAuthflowV2HeaderHTML,
	TemplateWebAuthflowV2DividerHTML,
	TemplateWebAuthflowV2AlertMessageHTML,
	TemplateWebAuthflowV2OTPInputHTML,
	TemplateWebAuthflowV2PasswordInputHTML,
	TemplateWebAuthflowV2PasswordStrengthMeterHTML,
	TemplateWebAuthflowV2PasswordFieldHTML,
	TemplateWebAuthflowV2NewPasswordFieldHTML,
	TemplateWebAuthflowV2PhoneInputHTML,
	TemplateWebAuthflowV2ErrorHTML,
	TemplateWebAuthflowV2PasswordPolicyHTML,
	TemplateWebAuthflowV2BranchHTML,
	TemplateWebAuthflowV2LockoutHTML,
	TemplateWebAuthflowV2ForgotPasswordAlternativesHTML,
	TemplateWebAuthflowV2ErrorPageLayoutHTML,
	TemplateWebAuthflowV2DeviceTokenCheckboxHTML,
	TemplateWebAuthflowV2TermsOfServiceAndPrivacyPolicyFooterHTML,
	TemplateWebAuthflowV2WatermarkHTML,
	TemplateWebAuthflowV2CSRFErrorPageLayoutHTML,
	TemplateWebAuthflowV2BrandLogoHTML,
	TemplateWebAuthflowV2SettingsHeaderHTML,
	TemplateWebAuthflowV2SettingsItemHTML,
	TemplateWebAuthflowV2NavBar,
	TemplateWebAuthflowV2DateInputHTML,
	TemplateWebAuthflowV2UserProfilePic,
	TemplateWebAuthflowV2SettingsRadioHTML,
	TemplateWebAuthflowV2SettingsTextInput,
	TemplateWebAuthflowV2LocaleInputHTML,
	TemplateWebAuthflowV2TimezoneInput,
	TemplateWebAuthflowV2CountryInput,
}

var ComponentsHTML []*template.HTML
var SettingsComponentsHTML []*template.HTML
