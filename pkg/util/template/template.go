package template

import (
	"github.com/Masterminds/sprig"
	messageformat "github.com/iawaknahc/gomessageformat"
)

var templateValidator = NewValidator(
	AllowRangeNode(true),
	AllowTemplateNode(true),
	AllowDeclaration(true),
	MaxDepth(15),
)

func MakeTemplateFuncMap() map[string]interface{} {
	var templateFuncMap = sprig.HermeticHtmlFuncMap()
	templateFuncMap[messageformat.TemplateRuntimeFuncName] = messageformat.TemplateRuntimeFunc
	templateFuncMap["makemap"] = MakeMap
	return templateFuncMap
}

var templateFuncMap = MakeTemplateFuncMap()
