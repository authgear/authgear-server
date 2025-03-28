package template

var privateTemplateValidator = NewValidator(
	AllowRangeNode(true),
	AllowTemplateNode(true),
	AllowDeclaration(true),
	AllowIdentifierNode(true),
	ForbidIdentifiers([]string{
		"print",
		"printf",
		"println",
	}),
	MaxDepth(99),
)

var publicTemplateValidator = NewValidator(
	AllowRangeNode(true),
	AllowTemplateNode(false),
	AllowDeclaration(true),
	AllowIdentifierNode(true),
	ForbidIdentifiers([]string{
		"print",
		"printf",
		"println",
		"call",
		"html",
		"js",
		"urlquery",
	}),
	MaxDepth(10),
)
