package graphql

import "github.com/authgear/authgear-server/pkg/util/graphqlutil"

var AppConfig = graphqlutil.NewJSONObjectScalar(
	"AppConfig",
	"The `AppConfig` scalar type represents an app config JSON object",
)

var FeatureConfig = graphqlutil.NewJSONObjectScalar(
	"FeatureConfig",
	"The `FeatureConfig` scalar type represents an feature config JSON object",
)

var AuthenticationFlowConfig = graphqlutil.NewJSONObjectScalar(
	"AuthenticationFlowConfig",
	"The `AuthenticationFlowConfig` scalar type represents the authentication_flow section of an app config",
)

var TutorialStatusData = graphqlutil.NewJSONObjectScalar(
	"TutorialStatusData",
	"The `TutorialStatusData` scalar type represents tutorial status data",
)

var ProjectWizardData = graphqlutil.NewJSONObjectScalar(
	"ProjectWizardData",
	"The `ProjectWizardData` scalar type represents form data of project wizard",
)

var StripeError = graphqlutil.NewJSONObjectScalar(
	"StripeError",
	"The `StripeError` scalar type represents Stripe error",
)
