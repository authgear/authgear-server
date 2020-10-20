package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateWebHTMLHeadHTML = template.RegisterHTML("web/__html_head.html")
var TemplateWebHeaderHTML = template.RegisterHTML("web/__header.html")
var TemplateWebNavBarHTML = template.RegisterHTML("web/__nav_bar.html")
var TemplateWebErrorHTML = template.RegisterHTML("web/__error.html")
var TemplateWebPasswordPolicyHTML = template.RegisterHTML("web/__password_policy.html")

var components = []*template.HTML{
	TemplateWebHTMLHeadHTML,
	TemplateWebHeaderHTML,
	TemplateWebNavBarHTML,
	TemplateWebErrorHTML,
	TemplateWebPasswordPolicyHTML,
}
