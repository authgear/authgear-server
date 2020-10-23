package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var createUserInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateUserInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"definition": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(identityDef),
			Description: "Definition of the identity of new user.",
		},
		"password": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Password for the user if required.",
		},
	},
})

var createUserPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateUserPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"createUser",
	&graphql.Field{
		Description: "Create new user",
		Type:        graphql.NewNonNull(createUserPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createUserInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			defData := input["definition"].(map[string]interface{})
			identityDef, err := model.ParseIdentityDef(defData)
			if err != nil {
				return nil, err
			}

			password, _ := input["password"].(string)

			gqlCtx := GQLContext(p.Context)

			userID, err := gqlCtx.UserFacade.Create(identityDef, password)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil
		},
	},
)

var resetPasswordInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "ResetPasswordInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"password": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "New password.",
		},
	},
})

var resetPasswordPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "ResetPasswordPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"resetPassword",
	&graphql.Field{
		Description: "Reset password of user",
		Type:        graphql.NewNonNull(resetPasswordPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(resetPasswordInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			password, _ := input["password"].(string)

			gqlCtx := GQLContext(p.Context)

			err := gqlCtx.UserFacade.ResetPassword(userID, password)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil
		},
	},
)

var setVerifiedStatusInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SetVerifiedStatusInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"claimName": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Name of the claim to set verified status.",
		},
		"claimValue": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Value of the claim.",
		},
		"isVerified": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Indicate whether the target claim is verified.",
		},
	},
})

var setVerifiedStatusPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "SetVerifiedStatusPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"setVerifiedStatus",
	&graphql.Field{
		Description: "Set verified status of a claim of user",
		Type:        graphql.NewNonNull(setVerifiedStatusPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(setVerifiedStatusInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userNodeID := input["userID"].(string)
			resolvedNodeID := relay.FromGlobalID(userNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeUser {
				return nil, apierrors.NewInvalid("invalid user ID")
			}
			userID := resolvedNodeID.ID

			claimName, _ := input["claimName"].(string)
			claimValue, _ := input["claimValue"].(string)
			isVerified, _ := input["isVerified"].(bool)

			gqlCtx := GQLContext(p.Context)

			err := gqlCtx.VerificationFacade.SetVerified(userID, claimName, claimValue, isVerified)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil
		},
	},
)
