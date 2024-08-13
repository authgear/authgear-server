package graphql

import (
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeUser = "User"

func init() {
	// Role and user, group and user forms a initialization cycle.
	// So we break the cycle by using AddFieldConfig.
	nodeUser.AddFieldConfig("roles", &graphql.Field{
		Type:        connRole.ConnectionType,
		Description: "The list of roles this user has.",
		Args:        relay.NewConnectionArgs(graphql.FieldConfigArgument{}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			source := p.Source.(*model.User)
			gqlCtx := GQLContext(p.Context)

			roles, err := gqlCtx.RolesGroupsFacade.ListRolesByUserID(source.ID)
			if err != nil {
				return nil, err
			}

			roleIfaces := make([]interface{}, len(roles))
			for i, r := range roles {
				roleIfaces[i] = r
			}

			args := relay.NewConnectionArguments(p.Args)
			return graphqlutil.NewConnectionFromArray(roleIfaces, args), nil
		},
	})

	nodeUser.AddFieldConfig("groups", &graphql.Field{
		Type:        connGroup.ConnectionType,
		Description: "The list of groups this user has.",
		Args:        relay.NewConnectionArgs(graphql.FieldConfigArgument{}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			source := p.Source.(*model.User)
			gqlCtx := GQLContext(p.Context)

			groups, err := gqlCtx.RolesGroupsFacade.ListGroupsByUserID(source.ID)
			if err != nil {
				return nil, err
			}

			groupIfaces := make([]interface{}, len(groups))
			for i, r := range groups {
				groupIfaces[i] = r
			}

			args := relay.NewConnectionArguments(p.Args)
			return graphqlutil.NewConnectionFromArray(groupIfaces, args), nil
		},
	})

	nodeUser.AddFieldConfig("effectiveRoles", &graphql.Field{
		Type:        connRole.ConnectionType,
		Description: "The list of computed roles this user has.",
		Args:        relay.NewConnectionArgs(graphql.FieldConfigArgument{}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			source := p.Source.(*model.User)
			gqlCtx := GQLContext(p.Context)

			roles, err := gqlCtx.RolesGroupsFacade.ListEffectiveRolesByUserID(source.ID)
			if err != nil {
				return nil, err
			}

			roleIfaces := make([]interface{}, len(roles))
			for i, r := range roles {
				roleIfaces[i] = r
			}

			args := relay.NewConnectionArguments(p.Args)
			return graphqlutil.NewConnectionFromArray(roleIfaces, args), nil
		},
	})
}

var nodeUser = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeUser,
		Description: "Authgear user",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeUser),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"lastLoginAt": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The last login time of user",
			},
			"loginIDs": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeIdentity))),
				Description: "The list of login ids",
				Resolve:     identitiesResolverByType(model.IdentityTypeLoginID),
			},
			"oauthConnections": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeIdentity))),
				Description: "The list of oauth connections",
				Resolve:     identitiesResolverByType(model.IdentityTypeOAuth),
			},
			"biometricRegistrations": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeIdentity))),
				Description: "The list of biometric registrations",
				Resolve:     identitiesResolverByType(model.IdentityTypeBiometric),
			},
			"passkeys": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeIdentity))),
				Description: "The list of passkeys",
				Resolve:     identitiesResolverByType(model.IdentityTypePasskey),
			},
			"identities": &graphql.Field{
				Type:        connIdentity.ConnectionType,
				Description: "The list of identities",
				Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
					"identityType": &graphql.ArgumentConfig{
						Type: identityType,
					},
				}),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)
					identityTypeStr, _ := p.Args["identityType"].(string)
					var identityTypePtr *model.IdentityType
					if identityTypeStr != "" {
						identityType := model.IdentityType(identityTypeStr)
						identityTypePtr = &identityType
					}

					refs, err := gqlCtx.IdentityFacade.List(source.ID, identityTypePtr)
					if err != nil {
						return nil, err
					}

					var identities []interface{}
					for _, i := range refs {
						identities = append(identities, i)
					}
					args := relay.NewConnectionArguments(p.Args)
					return graphqlutil.NewConnectionFromArray(identities, args), nil
				},
			},
			"primaryPassword": &graphql.Field{
				Type:        nodeAuthenticator,
				Description: "The primary password authenticator",
				Resolve:     authenticatorResolverByTypeAndKind(model.AuthenticatorTypePassword, authenticator.KindPrimary),
			},
			"primaryOOBOTPEmailAuthenticator": &graphql.Field{
				Type:        nodeAuthenticator,
				Description: "The primary passwordless via email authenticator",
				Resolve:     authenticatorResolverByTypeAndKind(model.AuthenticatorTypeOOBEmail, authenticator.KindPrimary),
			},
			"primaryOOBOTPSMSAuthenticator": &graphql.Field{
				Type:        nodeAuthenticator,
				Description: "The primary passwordless via phone authenticator",
				Resolve:     authenticatorResolverByTypeAndKind(model.AuthenticatorTypeOOBSMS, authenticator.KindPrimary),
			},
			"secondaryTOTPAuthenticators": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeAuthenticator))),
				Description: "The list of secondary TOTP authenticators",
				Resolve:     authenticatorsResolverByTypeAndKind(model.AuthenticatorTypeTOTP, authenticator.KindSecondary),
			},
			"secondaryOOBOTPEmailAuthenticators": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeAuthenticator))),
				Description: "The list of secondary passwordless via email authenticators",
				Resolve:     authenticatorsResolverByTypeAndKind(model.AuthenticatorTypeOOBEmail, authenticator.KindSecondary),
			},
			"secondaryOOBOTPSMSAuthenticators": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeAuthenticator))),
				Description: "The list of secondary passwordless via phone authenticators",
				Resolve:     authenticatorsResolverByTypeAndKind(model.AuthenticatorTypeOOBSMS, authenticator.KindSecondary),
			},
			"secondaryPassword": &graphql.Field{
				Type:        nodeAuthenticator,
				Description: "The secondary password authenticator",
				Resolve:     authenticatorResolverByTypeAndKind(model.AuthenticatorTypePassword, authenticator.KindSecondary),
			},
			"authenticators": &graphql.Field{
				Type:        connAuthenticator.ConnectionType,
				Description: "The list of authenticators",
				Args: relay.NewConnectionArgs(graphql.FieldConfigArgument{
					"authenticatorType": &graphql.ArgumentConfig{
						Type: authenticatorType,
					},
					"authenticatorKind": &graphql.ArgumentConfig{
						Type: authenticatorKind,
					},
				}),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)

					authenticatorTypeStr, _ := p.Args["authenticatorType"].(string)
					var authenticatorTypePtr *model.AuthenticatorType
					if authenticatorTypeStr != "" {
						authenticatorType := model.AuthenticatorType(authenticatorTypeStr)
						authenticatorTypePtr = &authenticatorType
					}

					authenticatorKindStr, _ := p.Args["authenticatorKind"].(string)
					var authenticatorKindPtr *authenticator.Kind
					if authenticatorKindStr != "" {
						authenticatorKind := authenticator.Kind(authenticatorKindStr)
						authenticatorKindPtr = &authenticatorKind
					}

					refs, err := gqlCtx.AuthenticatorFacade.List(source.ID, authenticatorTypePtr, authenticatorKindPtr)
					if err != nil {
						return nil, err
					}

					var authenticators []interface{}
					for _, i := range refs {
						authenticators = append(authenticators, i)
					}
					args := relay.NewConnectionArguments(p.Args)
					return graphqlutil.NewConnectionFromArray(authenticators, args), nil

				},
			},
			"verifiedClaims": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(claim))),
				Description: "The list of user's verified claims",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)
					claims, err := gqlCtx.VerificationFacade.Get(source.ID)
					if err != nil {
						return nil, err
					}

					return claims, nil
				},
			},
			"sessions": &graphql.Field{
				Type:        connSession.ConnectionType,
				Description: "The list of first party app sessions",
				Args:        relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)
					ss, err := gqlCtx.SessionFacade.List(source.ID)
					if err != nil {
						return nil, err
					}

					sessionModels, err := gqlCtx.SessionListing.FilterForDisplay(ss, nil)
					if err != nil {
						return nil, err
					}

					var sessions []interface{}
					for _, i := range sessionModels {
						sessions = append(sessions, i.Session)
					}
					args := relay.NewConnectionArguments(p.Args)
					return graphqlutil.NewConnectionFromArray(sessions, args), nil
				},
			},
			"authorizations": &graphql.Field{
				Type:        connAuthorization.ConnectionType,
				Description: "The list of third party app authorizations",
				Args:        relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)

					// return third party client authorizations only in admin api
					filter := oauth.NewKeepThirdPartyAuthorizationFilter(gqlCtx.OAuthConfig)
					as, err := gqlCtx.AuthorizationFacade.List(source.ID, filter)
					if err != nil {
						return nil, err
					}

					var authzs []interface{}
					for _, i := range as {
						authzs = append(authzs, i.ToAPIModel())
					}
					args := relay.NewConnectionArguments(p.Args)
					return graphqlutil.NewConnectionFromArray(authzs, args), nil
				},
			},
			"isDisabled": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Boolean),
				Description: "Indicates if the user is disabled",
			},
			"isAnonymous": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Boolean),
				Description: "Indicates if the user is anonymous",
			},
			"disableReason": &graphql.Field{
				Type:        graphql.String,
				Description: "The reason of disabled",
			},
			"isDeactivated": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Boolean),
				Description: "Indicates if the user is deactivated",
			},
			"deleteAt": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The scheduled deletion time of the user",
			},
			"isAnonymized": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Boolean),
				Description: "Indicates if the user is anonymized",
			},
			"anonymizeAt": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The scheduled anonymization time of the user",
			},
			"standardAttributes": &graphql.Field{
				Type:        graphql.NewNonNull(UserStandardAttributes),
				Description: "The user's standard attributes",
			},
			"customAttributes": &graphql.Field{
				Type:        graphql.NewNonNull(UserCustomAttributes),
				Description: "The user's custom attributes",
			},
			"web3": &graphql.Field{
				Type:        graphql.NewNonNull(Web3Claims),
				Description: "The web3 claims",
			},
			"formattedName": &graphql.Field{
				Type:        graphql.String,
				Description: "The user's formatted name",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					formattedName := stdattrs.T(source.StandardAttributes).FormattedName()
					if formattedName == "" {
						return nil, nil
					}
					return formattedName, nil
				},
			},
			"endUserAccountID": &graphql.Field{
				Type:        graphql.String,
				Description: "The end user account id constructed based on user's personal data. (e.g. email, phone...etc)",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					endUserAccountID := source.EndUserAccountID
					if endUserAccountID == "" {
						return nil, nil
					}
					return endUserAccountID, nil
				},
			},
			"mfaGracePeriodEndAt": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "Indicate when will user's MFA grace period will end",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					return source.MFAGracePeriodtEndAt, nil
				},
			},
		},
	}),
	&model.User{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Users.Load(id).Value, nil
	},
)

var connUser = graphqlutil.NewConnectionDef(nodeUser)

func identitiesResolverByType(typ model.IdentityType) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		source := p.Source.(*model.User)
		gqlCtx := GQLContext(p.Context)
		identities, err := gqlCtx.IdentityFacade.List(source.ID, &typ)
		if err != nil {
			return nil, err
		}

		return identities, nil
	}
}

func authenticatorResolverByTypeAndKind(authenticatorType apimodel.AuthenticatorType, authenticatorKind authenticator.Kind) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		source := p.Source.(*model.User)
		gqlCtx := GQLContext(p.Context)

		authenticators, err := gqlCtx.AuthenticatorFacade.List(source.ID, &authenticatorType, &authenticatorKind)
		if err != nil {
			return nil, err
		}

		if len(authenticators) > 0 {
			return authenticators[0], nil
		}

		return nil, nil
	}
}

func authenticatorsResolverByTypeAndKind(authenticatorType apimodel.AuthenticatorType, authenticatorKind authenticator.Kind) func(p graphql.ResolveParams) (interface{}, error) {
	return func(p graphql.ResolveParams) (interface{}, error) {
		source := p.Source.(*model.User)
		gqlCtx := GQLContext(p.Context)

		authenticators, err := gqlCtx.AuthenticatorFacade.List(source.ID, &authenticatorType, &authenticatorKind)
		if err != nil {
			return nil, err
		}

		return authenticators, nil
	}
}
