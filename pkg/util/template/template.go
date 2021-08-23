package template

var templateValidator = NewValidator(
	AllowRangeNode(true),
	AllowTemplateNode(true),
	AllowDeclaration(true),
	MaxDepth(99),
)
