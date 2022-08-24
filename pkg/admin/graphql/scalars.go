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

var AuditLogData = graphqlutil.NewJSONObjectScalar(
	"AuditLogData",
	"The `AuditLogData` scalar type represents the data of the audit log",
)

var UserStandardAttributes = graphqlutil.NewJSONObjectScalar(
	"UserStandardAttributes",
	"The `UserStandardAttributes` scalar type represents the standard attributes of the user",
)

var UserCustomAttributes = graphqlutil.NewJSONObjectScalar(
	"UserCustomAttributes",
	"The `UserCustomAttributes` scalar type represents the custom attributes of the user",
)

var Web3Claims = graphqlutil.NewJSONObjectScalar(
	"Web3Claims",
	"The `Web3Claims` scalar type represents the scalar type of the user",
)
