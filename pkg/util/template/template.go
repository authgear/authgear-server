package template

import (
	messageformat "github.com/iawaknahc/gomessageformat"
)

var templateValidator = NewValidator(
	AllowRangeNode(true),
	AllowTemplateNode(true),
	AllowDeclaration(true),
	MaxDepth(15),
)

var templateFuncMap = map[string]interface{}{
	messageformat.TemplateRuntimeFuncName: messageformat.TemplateRuntimeFunc,
	"makemap":                             MakeMap,
}
