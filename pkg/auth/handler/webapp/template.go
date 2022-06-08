package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

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

var components = []*template.HTML{
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
}

var TemplateWebRecoveryCodeTXT = template.RegisterPlainText("web/__recovery_code.txt")

var plainTextComponents = []*template.PlainText{
	TemplateWebRecoveryCodeTXT,
}
