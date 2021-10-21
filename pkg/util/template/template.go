package template

var templateValidator = NewValidator(
	AllowRangeNode(true),
	AllowTemplateNode(true),
	AllowDeclaration(true),
	AllowIdentifierNode(true),
	MaxDepth(99),
)
