package graphql

import (
	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var deleteAuthenticatorInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteAuthenticatorInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"authenticatorID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target authenticator ID.",
		},
	},
})

var deleteAuthenticatorPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteAuthenticatorPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"deleteAuthenticator",
	&graphql.Field{
		Description: "Delete authenticator of user",
		Type:        graphql.NewNonNull(deleteAuthenticatorPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteAuthenticatorInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			authenticatorNodeID := input["authenticatorID"].(string)

			resolvedNodeID := relay.FromGlobalID(authenticatorNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeAuthenticator {
				return nil, apierrors.NewInvalid("invalid authenticator ID")
			}

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			info, err := gqlCtx.AuthenticatorFacade.Get(ctx, resolvedNodeID.ID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuthenticatorFacade.Remove(ctx, info)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationDeleteAuthenticatorExecutedEventPayload{
				Authenticator: info.ToModel(),
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: info.UserID,
					},
				},
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(ctx, info.UserID),
			}).Value, nil
		},
	},
)

var createAuthenticatorInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateAuthenticatorInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"definition": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(authenticatorDef),
			Description: "Definition of the new authenticator.",
		},
	},
})

var createAuthenticatorPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateAuthenticatorPayload",
	Fields: graphql.Fields{
		"authenticator": &graphql.Field{
			Type: graphql.NewNonNull(nodeAuthenticator),
		},
	},
})

var _ = registerMutationField(
	"createAuthenticator",
	&graphql.Field{
		Description: "Create authenticator of user",
		Type:        graphql.NewNonNull(createAuthenticatorPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createAuthenticatorInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			definition := input["definition"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			kind := definition["kind"].(string)
			authnType := definition["type"].(string)

			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			spec := &authenticator.Spec{
				UserID: userID,
				Kind:   apimodel.AuthenticatorKind(kind),
			}

			switch authnType {
			case string(apimodel.AuthenticatorTypeOOBEmail):
				oobOtpEmail, ok := definition["oobOtpEmail"].(map[string]interface{})
				if !ok {
					return nil, apierrors.NewInvalid("definition/oobOtpEmail is required")
				}
				spec.Type = apimodel.AuthenticatorTypeOOBEmail
				spec.OOBOTP = &authenticator.OOBOTPSpec{
					Email: oobOtpEmail["email"].(string),
				}
			case string(apimodel.AuthenticatorTypeOOBSMS):
				oobOtpSMS, ok := definition["oobOtpSMS"].(map[string]interface{})
				if !ok {
					return nil, apierrors.NewInvalid("definition/oobOtpSMS is required")
				}
				spec.Type = apimodel.AuthenticatorTypeOOBSMS
				spec.OOBOTP = &authenticator.OOBOTPSpec{
					Phone: oobOtpSMS["phone"].(string),
				}
			case string(apimodel.AuthenticatorTypePassword):
				password, ok := definition["password"].(map[string]interface{})
				if !ok {
					return nil, apierrors.NewInvalid("definition/password is required")
				}
				spec.Type = apimodel.AuthenticatorTypePassword
				spec.Password = &authenticator.PasswordSpec{
					PlainPassword: password["password"].(string),
				}
			default:
				return nil, apierrors.NewInvalid("unsupported authenticator type")
			}

			info, err := gqlCtx.AuthenticatorFacade.CreateBySpec(ctx, spec)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(ctx, &nonblocking.AdminAPIMutationCreateAuthenticatorExecutedEventPayload{
				Authenticator: info.ToModel(),
				UserRef: apimodel.UserRef{
					Meta: apimodel.Meta{
						ID: info.UserID,
					},
				},
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"authenticator": gqlCtx.Authenticators.Load(ctx, info.ID),
			}).Value, nil
		},
	},
)
