package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
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

var setDisabledStatusInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SetDisabledStatusInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"isDisabled": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Indicate whether the target user is disabled.",
		},
		"reason": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Indicate the disable reason; If not provided, the user will be disabled with no reason.",
		},
	},
})

var setDisabledStatusPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "SetDisabledStatusPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"setDisabledStatus",
	&graphql.Field{
		Description: "Set disabled status of user",
		Type:        graphql.NewNonNull(setDisabledStatusPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(setDisabledStatusInput),
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

			isDisabled := input["isDisabled"].(bool)
			var reason *string
			if r, ok := input["reason"].(string); ok && isDisabled {
				reason = &r
			}

			gqlCtx := GQLContext(p.Context)

			err := gqlCtx.UserFacade.SetDisabled(userID, isDisabled, reason)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil
		},
	},
)

var updateUserInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateUserInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"standardAttributes": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(UserStandardAttributes),
			Description: "Whole standard attributes to be set on the user.",
		},
	},
})

var updateUserPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateUserPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"updateUser",
	&graphql.Field{
		Description: "Update user",
		Type:        graphql.NewNonNull(updateUserPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateUserInput),
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

			gqlCtx := GQLContext(p.Context)

			stdAttrs := input["standardAttributes"].(map[string]interface{})

			err := gqlCtx.UserFacade.UpdateStandardAttributes(accesscontrol.EmptyRole, userID, stdAttrs)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil
		},
	},
)

var deleteUserInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteUserInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
	},
})

var deleteUserPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteUserPayload",
	Fields: graphql.Fields{
		"deletedUserID": &graphql.Field{
			Type: graphql.NewNonNull(graphql.ID),
		},
	},
})

var _ = registerMutationField(
	"deleteUser",
	&graphql.Field{
		Description: "Delete specified user",
		Type:        graphql.NewNonNull(deleteUserPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteUserInput),
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

			gqlCtx := GQLContext(p.Context)

			err := gqlCtx.UserFacade.Delete(userID)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"deletedUserID": userNodeID,
			}, nil
		},
	},
)
