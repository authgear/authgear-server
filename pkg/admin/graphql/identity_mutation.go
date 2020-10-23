package graphql

import (
	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/admin/loader"
	"github.com/authgear/authgear-server/pkg/admin/model"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var deleteIdentityInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteIdentityInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"identityID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target identity ID.",
		},
	},
})

var deleteIdentityPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteIdentityPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"deleteIdentity",
	&graphql.Field{
		Description: "Delete identity of user",
		Type:        graphql.NewNonNull(deleteIdentityPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteIdentityInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})
			identityNodeID := input["identityID"].(string)

			resolvedNodeID := relay.FromGlobalID(identityNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeIdentity {
				return nil, apierrors.NewInvalid("invalid identity ID")
			}
			identityRef, err := loader.DecodeIdentityID(resolvedNodeID.ID)
			if err != nil {
				return nil, apierrors.NewInvalid("invalid identity ID")
			}

			gqlCtx := GQLContext(p.Context)

			info, err := gqlCtx.IdentityFacade.Get(identityRef)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.IdentityFacade.Remove(info)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(info.UserID),
			}).Value, nil
		},
	},
)

var createIdentityInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateIdentityInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target user ID.",
		},
		"definition": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(identityDef),
			Description: "Definition of the new identity.",
		},
		"password": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "Password for the user if required.",
		},
	},
})

var createIdentityPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateIdentityPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
		"identity": &graphql.Field{
			Type: graphql.NewNonNull(nodeIdentity),
		},
	},
})

var _ = registerMutationField(
	"createIdentity",
	&graphql.Field{
		Description: "Create new identity for user",
		Type:        graphql.NewNonNull(createIdentityPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createIdentityInput),
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

			defData := input["definition"].(map[string]interface{})
			identityDef, err := model.ParseIdentityDef(defData)
			if err != nil {
				return nil, err
			}

			password, _ := input["password"].(string)

			gqlCtx := GQLContext(p.Context)

			ref, err := gqlCtx.IdentityFacade.Create(userID, identityDef, password)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user":     gqlCtx.Users.Load(userID),
				"identity": ref,
			}).Value, nil
		},
	},
)
