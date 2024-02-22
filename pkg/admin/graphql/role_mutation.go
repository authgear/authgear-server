package graphql

import (
	relay "github.com/authgear/graphql-go-relay"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/graphql-go/graphql"
)

var createRoleInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateRoleInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"key": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the role.",
		},
		"name": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The optional name of the role.",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The optional description of the role.",
		},
	},
})

var createRolePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateRolePayload",
	Fields: graphql.Fields{
		"role": &graphql.Field{
			Type: graphql.NewNonNull(nodeRole),
		},
	},
})

var _ = registerMutationField(
	"createRole",
	&graphql.Field{
		Description: "Create a new role.",
		Type:        graphql.NewNonNull(createRolePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createRoleInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			key := input["key"].(string)

			var name *string
			if str, ok := input["name"].(string); ok && str != "" {
				name = &str
			}

			var description *string
			if str, ok := input["description"].(string); ok && str != "" {
				description = &str
			}

			options := &rolesgroups.NewRoleOptions{
				Key:         key,
				Name:        name,
				Description: description,
			}

			gqlCtx := GQLContext(p.Context)
			roleID, err := gqlCtx.RolesGroupsFacade.CreateRole(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"role": gqlCtx.Roles.Load(roleID),
			}).Value, nil
		},
	},
)

var updateRoleInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateRoleInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the role.",
		},
		"key": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The new key of the role. Pass null if you do not need to update the key.",
		},
		"name": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The new name of the role. Pass null if you do not need to update the name. Pass an empty string to remove the name.",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The new description of the role. Pass null if you do not need to update the description. Pass an empty string to remove the description.",
		},
	},
})

var updateRolePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateRolePayload",
	Fields: graphql.Fields{
		"role": &graphql.Field{
			Type: graphql.NewNonNull(nodeRole),
		},
	},
})

var _ = registerMutationField(
	"updateRole",
	&graphql.Field{
		Description: "Update an existing role.",
		Type:        graphql.NewNonNull(updateRolePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateRoleInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			roleNodeID := input["id"].(string)

			resolvedNodeID := relay.FromGlobalID(roleNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeRole {
				return nil, apierrors.NewInvalid("invalid role ID")
			}
			roleID := resolvedNodeID.ID

			var newKey *string
			if str, ok := input["key"].(string); ok {
				newKey = &str
			}

			var newName *string
			if str, ok := input["name"].(string); ok {
				newName = &str
			}

			var newDescription *string
			if str, ok := input["description"].(string); ok {
				newDescription = &str
			}

			options := &rolesgroups.UpdateRoleOptions{
				ID:             roleID,
				NewKey:         newKey,
				NewName:        newName,
				NewDescription: newDescription,
			}

			gqlCtx := GQLContext(p.Context)
			err := gqlCtx.RolesGroupsFacade.UpdateRole(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"role": gqlCtx.Roles.Load(roleID),
			}).Value, nil
		},
	},
)

var deleteRoleInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteRoleInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the role.",
		},
	},
})

var deleteRolePayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteRolePayload",
	Fields: graphql.Fields{
		"ok": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

var _ = registerMutationField(
	"deleteRole",
	&graphql.Field{
		Description: "Delete an existing role. The associations between the role with other groups and other users will also be deleted.",
		Type:        graphql.NewNonNull(deleteRolePayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteRoleInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			roleNodeID := input["id"].(string)

			resolvedNodeID := relay.FromGlobalID(roleNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeRole {
				return nil, apierrors.NewInvalid("invalid role ID")
			}
			roleID := resolvedNodeID.ID

			gqlCtx := GQLContext(p.Context)
			err := gqlCtx.RolesGroupsFacade.DeleteRole(roleID)
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"ok": true,
			}, nil
		},
	},
)
