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

var AuthenticationFlowSignupFlow = graphqlutil.NewJSONObjectScalar(
	"AuthenticationFlowSignupFlow",
	"The `AuthenticationFlowSignupFlow` scalar type represents an item of authentication_flow.signup_flows in an app config",
)

var AuthenticationFlowLoginFlow = graphqlutil.NewJSONObjectScalar(
	"AuthenticationFlowLoginFlow",
	"The `AuthenticationFlowLoginFlow` scalar type represents an item of authentication_flow.login_flows in an app config",
)

var AuthenticationFlowSignupLoginFlow = graphqlutil.NewJSONObjectScalar(
	"AuthenticationFlowSignupLoginFlow",
	"The `AuthenticationFlowSignupLoginFlow` scalar type represents an item of authentication_flow.signup_login_flows in an app config",
)

var AuthenticationFlowReauthFlow = graphqlutil.NewJSONObjectScalar(
	"AuthenticationFlowReauthFlow",
	"The `AuthenticationFlowReauthFlow` scalar type represents an item of authentication_flow.reauth_flows in an app config",
)

var AuthenticationFlowAccountRecoveryFlow = graphqlutil.NewJSONObjectScalar(
	"AuthenticationFlowAccountRecoveryFlow",
	"The `AuthenticationFlowAccountRecoveryFlow` scalar type represents an item of authentication_flow.account_recovery_flows in an app config",
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
