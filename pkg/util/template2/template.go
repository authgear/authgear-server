package template

import (
	messageformat "github.com/iawaknahc/gomessageformat"

	"github.com/authgear/authgear-server/pkg/util/template"
)

var templateValidator = template.NewValidator(
	template.AllowRangeNode(true),
	template.AllowTemplateNode(true),
	template.AllowDeclaration(true),
	template.MaxDepth(15),
)

var templateFuncMap = map[string]interface{}{
	messageformat.TemplateRuntimeFuncName: messageformat.TemplateRuntimeFunc,
	"makemap":                             template.MakeMap,
}
