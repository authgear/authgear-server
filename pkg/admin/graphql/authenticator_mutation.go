package graphql

import (
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

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

			gqlCtx := GQLContext(p.Context)

			info, err := gqlCtx.AuthenticatorFacade.Get(resolvedNodeID.ID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuthenticatorFacade.Remove(info)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(&nonblocking.AdminAPIMutationDeleteAuthenticatorExecutedEventPayload{
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
				"user": gqlCtx.Users.Load(info.UserID),
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

			gqlCtx := GQLContext(p.Context)

			spec := &authenticator.Spec{
				UserID: userID,
			}

			if oobOtpEmail, ok := definition["oobOtpEmail"].(map[string]interface{}); ok && oobOtpEmail != nil {
				spec.Type = apimodel.AuthenticatorTypeOOBEmail
				spec.OOBOTP = &authenticator.OOBOTPSpec{
					Email: oobOtpEmail["email"].(string),
				}
				spec.Kind = apimodel.AuthenticatorKind(oobOtpEmail["kind"].(string))
			} else if oobOtpSMS, ok := definition["oobOtpSMS"].(map[string]interface{}); ok && oobOtpSMS != nil {
				spec.Type = apimodel.AuthenticatorTypeOOBSMS
				spec.OOBOTP = &authenticator.OOBOTPSpec{
					Phone: oobOtpSMS["phone"].(string),
				}
				spec.Kind = apimodel.AuthenticatorKind(oobOtpSMS["kind"].(string))
			} else if password, ok := definition["password"].(map[string]interface{}); ok && password != nil {
				spec.Type = apimodel.AuthenticatorTypePassword
				spec.Password = &authenticator.PasswordSpec{
					PlainPassword: password["password"].(string),
				}
				spec.Kind = apimodel.AuthenticatorKind(password["kind"].(string))
			} else {
				panic("unsupported authenticator type")
			}

			info, err := gqlCtx.AuthenticatorFacade.CreateBySpec(spec)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(&nonblocking.AdminAPIMutationCreateAuthenticatorExecutedEventPayload{
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
				"authenticator": gqlCtx.Authenticators.Load(info.ID),
			}).Value, nil
		},
	},
)
