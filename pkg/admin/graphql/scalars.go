package graphql

import "github.com/authgear/authgear-server/pkg/util/graphqlutil"

var IdentityClaims = graphqlutil.NewJSONObjectScalar(
	"IdentityClaims",
	"The `IdentityClaims` scalar type represents a set of claims belonging to an identity",
)

var AuthenticatorClaims = graphqlutil.NewJSONObjectScalar(
	"AuthenticatorClaims",
	"The `AuthenticatorClaims` scalar type represents a set of claims belonging to an authenticator",
)
