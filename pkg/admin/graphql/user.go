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
				Type:    graphql.NewList(graphql.NewNonNull(nodeIdentity)),
				Resolve: identitiesResolverByType(model.IdentityTypeLoginID),
			},
			"oauthConnections": &graphql.Field{
				Type:    graphql.NewList(graphql.NewNonNull(nodeIdentity)),
				Resolve: identitiesResolverByType(model.IdentityTypeOAuth),
			},
			"biometricRegistrations": &graphql.Field{
				Type:    graphql.NewList(graphql.NewNonNull(nodeIdentity)),
				Resolve: identitiesResolverByType(model.IdentityTypeBiometric),
			},
			"passkeys": &graphql.Field{
				Type:    graphql.NewList(graphql.NewNonNull(nodeIdentity)),
				Resolve: identitiesResolverByType(model.IdentityTypePasskey),
			},
			"identities": &graphql.Field{
				Type: connIdentity.ConnectionType,
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
				Type:    nodeAuthenticator,
				Resolve: authenticatorResolverByTypeAndKind(model.AuthenticatorTypePassword, authenticator.KindPrimary),
			},
			"primaryOOBOTPEmailAuthenticators": &graphql.Field{
				Type:    graphql.NewList(graphql.NewNonNull(nodeAuthenticator)),
				Resolve: authenticatorsResolverByTypeAndKind(model.AuthenticatorTypeOOBEmail, authenticator.KindPrimary),
			},
			"primaryOOBOTPSMSAuthenticators": &graphql.Field{
				Type:    graphql.NewList(graphql.NewNonNull(nodeAuthenticator)),
				Resolve: authenticatorsResolverByTypeAndKind(model.AuthenticatorTypeOOBSMS, authenticator.KindPrimary),
			},
			"secondaryTOTPAuthenticators": &graphql.Field{
				Type:    graphql.NewList(graphql.NewNonNull(nodeAuthenticator)),
				Resolve: authenticatorsResolverByTypeAndKind(model.AuthenticatorTypeTOTP, authenticator.KindSecondary),
			},
			"secondaryOOBOTPEmailAuthenticators": &graphql.Field{
				Type:    graphql.NewList(graphql.NewNonNull(nodeAuthenticator)),
				Resolve: authenticatorsResolverByTypeAndKind(model.AuthenticatorTypeOOBEmail, authenticator.KindSecondary),
			},
			"secondaryOOBOTPSMSAuthenticators": &graphql.Field{
				Type:    graphql.NewList(graphql.NewNonNull(nodeAuthenticator)),
				Resolve: authenticatorsResolverByTypeAndKind(model.AuthenticatorTypeOOBSMS, authenticator.KindSecondary),
			},
			"secondaryPassword": &graphql.Field{
				Type:    nodeAuthenticator,
				Resolve: authenticatorResolverByTypeAndKind(model.AuthenticatorTypePassword, authenticator.KindSecondary),
			},
			"authenticators": &graphql.Field{
				Type: connAuthenticator.ConnectionType,
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
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(claim))),
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
				Type: connSession.ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					gqlCtx := GQLContext(p.Context)
					ss, err := gqlCtx.SessionFacade.List(source.ID)
					if err != nil {
						return nil, err
					}

					filter := oauth.NewRemoveThirdPartySessionFilter(gqlCtx.OAuthConfig)
					ss = oauth.ApplySessionFilters(ss, filter)

					var sessions []interface{}
					for _, i := range ss {
						sessions = append(sessions, i.ToAPIModel())
					}
					args := relay.NewConnectionArguments(p.Args)
					return graphqlutil.NewConnectionFromArray(sessions, args), nil
				},
			},
			"authorizations": &graphql.Field{
				Type: connAuthorization.ConnectionType,
				Args: relay.ConnectionArgs,
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
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"isAnonymous": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"disableReason": &graphql.Field{
				Type: graphql.String,
			},
			"isDeactivated": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
			},
			"deleteAt": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The scheduled deletion time of the user",
			},
			"standardAttributes": &graphql.Field{
				Type: graphql.NewNonNull(UserStandardAttributes),
			},
			"customAttributes": &graphql.Field{
				Type: graphql.NewNonNull(UserCustomAttributes),
			},
			"web3": &graphql.Field{
				Type: graphql.NewNonNull(Web3Claims),
			},
			"formattedName": &graphql.Field{
				Type: graphql.String,
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
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					source := p.Source.(*model.User)
					endUserAccountID := source.EndUserAccountID()
					if endUserAccountID == "" {
						return nil, nil
					}
					return endUserAccountID, nil
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
