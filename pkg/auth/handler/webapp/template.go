package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebHTMLHeadHTML = template.RegisterHTML("web/__html_head.html")
var TemplateWebHeaderHTML = template.RegisterHTML("web/__header.html")
var TemplateWebNavBarHTML = template.RegisterHTML("web/__nav_bar.html")
var TemplateWebErrorHTML = template.RegisterHTML("web/__error.html")
var TemplateWebMessageBarHTML = template.RegisterHTML("web/__message_bar.html")
var TemplateWebAlternativeStepsHTML = template.RegisterHTML("web/__alternatives.html")
var TemplateWebPasswordPolicyHTML = template.RegisterHTML("web/__password_policy.html")
var TemplateWebPageFrameHTML = template.RegisterHTML("web/__page_frame.html")
var TemplateWebModalHTML = template.RegisterHTML("web/__modal.html")

var components = []*template.HTML{
	TemplateWebHTMLHeadHTML,
	TemplateWebHeaderHTML,
	TemplateWebNavBarHTML,
	TemplateWebErrorHTML,
	TemplateWebMessageBarHTML,
	TemplateWebAlternativeStepsHTML,
	TemplateWebPasswordPolicyHTML,
	TemplateWebPageFrameHTML,
	TemplateWebModalHTML,
}
