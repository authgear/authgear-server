package template

var privateTemplateValidator = NewValidator(
	AllowRangeNode(true),
	AllowTemplateNode(true),
	AllowDeclaration(true),
	AllowIdentifierNode(true),
	MaxDepth(99),
)

var publicTemplateValidator = NewValidator(
	AllowRangeNode(true),
	AllowTemplateNode(false),
	AllowDeclaration(true),
	AllowIdentifierNode(true),
	MaxDepth(10),
)
