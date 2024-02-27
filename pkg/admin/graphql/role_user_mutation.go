package graphql

import (
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/graphql-go/graphql"
)

var addRoleToUsersInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AddRoleToUsersInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"roleKey": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the role.",
		},
		"userIDs": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.ID)),
			Description: "The list of user ids.",
		},
	},
})

var addRoleToUsersPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AddRoleToUsersPayload",
	Fields: graphql.Fields{
		"role": &graphql.Field{
			Type: graphql.NewNonNull(nodeRole),
		},
	},
})

var _ = registerMutationField(
	"addRoleToUsers",
	&graphql.Field{
		Description: "Add the role to the users.",
		Type:        graphql.NewNonNull(addRoleToUsersPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(addRoleToUsersInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			roleKey := input["roleKey"].(string)
			userIDIfaces := input["userIDs"].([]interface{})
			userIDs := make([]string, len(userIDIfaces))
			for i, v := range userIDIfaces {
				userIDs[i] = v.(string)
			}
			gqlCtx := GQLContext(p.Context)

			options := &rolesgroups.AddRoleToUsersOptions{
				RoleKey: roleKey,
				UserIDs: userIDs,
			}
			roleID, err := gqlCtx.RolesGroupsFacade.AddRoleToUsers(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"role": gqlCtx.Roles.Load(roleID),
			}).Value, nil

		},
	},
)

var removeRoleFromUsersInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RemoveRoleFromUsersInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"roleKey": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the role.",
		},
		"userIDs": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.ID)),
			Description: "The list of user ids.",
		},
	},
})

var removeRoleFromUsersPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RemoveRoleFromUsersPayload",
	Fields: graphql.Fields{
		"role": &graphql.Field{
			Type: graphql.NewNonNull(nodeRole),
		},
	},
})

var _ = registerMutationField(
	"removeRoleFromUsers",
	&graphql.Field{
		Description: "Remove the role to the users.",
		Type:        graphql.NewNonNull(removeRoleFromUsersPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(removeRoleFromUsersInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			roleKey := input["roleKey"].(string)
			userIDIfaces := input["userIDs"].([]interface{})
			userIDs := make([]string, len(userIDIfaces))
			for i, v := range userIDIfaces {
				userIDs[i] = v.(string)
			}
			gqlCtx := GQLContext(p.Context)

			options := &rolesgroups.RemoveRoleFromUsersOptions{
				RoleKey: roleKey,
				UserIDs: userIDs,
			}
			roleID, err := gqlCtx.RolesGroupsFacade.RemoveRoleFromUsers(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"role": gqlCtx.Roles.Load(roleID),
			}).Value, nil

		},
	},
)

var addUserToRolesInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AddUserToRolesInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The id of the user.",
		},
		"roleKeys": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.String)),
			Description: "The list of role keys.",
		},
	},
})

var addUserToRolesPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AddUserToRolesPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"addUserToRoles",
	&graphql.Field{
		Description: "Add the user to the roles.",
		Type:        graphql.NewNonNull(addUserToRolesPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(addUserToRolesInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userID := input["userID"].(string)
			roleKeyIfaces := input["roleKeys"].([]interface{})
			roleKeys := make([]string, len(roleKeyIfaces))
			for i, v := range roleKeyIfaces {
				roleKeys[i] = v.(string)
			}
			gqlCtx := GQLContext(p.Context)

			options := &rolesgroups.AddUserToRolesOptions{
				UserID:   userID,
				RoleKeys: roleKeys,
			}
			err := gqlCtx.RolesGroupsFacade.AddUserToRoles(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil

		},
	},
)

var removeUserFromRolesInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RemoveUserFromRolesInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"userID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The id of the user.",
		},
		"roleKeys": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.String)),
			Description: "The list of role keys.",
		},
	},
})

var removeUserFromRolesPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RemoveUserFromRolesPayload",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
		},
	},
})

var _ = registerMutationField(
	"removeUserFromRoles",
	&graphql.Field{
		Description: "Remove the user from the roles.",
		Type:        graphql.NewNonNull(removeUserFromRolesPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(removeUserFromRolesInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			userID := input["userID"].(string)
			roleKeyIfaces := input["roleKeys"].([]interface{})
			roleKeys := make([]string, len(roleKeyIfaces))
			for i, v := range roleKeyIfaces {
				roleKeys[i] = v.(string)
			}
			gqlCtx := GQLContext(p.Context)

			options := &rolesgroups.RemoveUserFromRolesOptions{
				UserID:   userID,
				RoleKeys: roleKeys,
			}
			err := gqlCtx.RolesGroupsFacade.RemoveUserFromRoles(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"user": gqlCtx.Users.Load(userID),
			}).Value, nil

		},
	},
)
