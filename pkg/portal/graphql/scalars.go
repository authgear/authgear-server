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
