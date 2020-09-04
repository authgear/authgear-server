package graphql

import "github.com/authgear/authgear-server/pkg/util/graphqlutil"

var AppConfig = graphqlutil.NewJSONObjectScalar(
	"AppConfig",
	"The `AppConfig` scalar type represents an app config JSON object",
)

var SecretConfig = graphqlutil.NewJSONObjectScalar(
	"SecretConfig",
	"The `SecretConfig` scalar type represents a secret config JSON object",
)
