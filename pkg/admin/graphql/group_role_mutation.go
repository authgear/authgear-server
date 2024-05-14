package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

var addRoleToGroupsInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AddRoleToGroupsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"roleKey": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the role.",
		},
		"groupKeys": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "The list of group keys.",
		},
	},
})

var addRoleToGroupsPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AddRoleToGroupsPayload",
	Fields: graphql.Fields{
		"role": &graphql.Field{
			Type: graphql.NewNonNull(nodeRole),
		},
	},
})

var _ = registerMutationField(
	"addRoleToGroups",
	&graphql.Field{
		Description: "Add the role to the groups.",
		Type:        graphql.NewNonNull(addRoleToGroupsPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(addRoleToGroupsInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			roleKey := input["roleKey"].(string)

			groupKeyIfaces := input["groupKeys"].([]interface{})
			groupKeys := make([]string, len(groupKeyIfaces))
			for i, v := range groupKeyIfaces {
				groupKeys[i] = v.(string)
			}

			options := &rolesgroups.AddRoleToGroupsOptions{
				RoleKey:   roleKey,
				GroupKeys: groupKeys,
			}

			gqlCtx := GQLContext(p.Context)

			affectedUserIDs, err := gqlCtx.RolesGroupsFacade.ListAllUserIDsByGroupKeys(groupKeys)
			if err != nil {
				return nil, err
			}

			groups, err := gqlCtx.RolesGroupsFacade.ListAllGroupsByKeys(groupKeys)
			if err != nil {
				return nil, err
			}

			roleID, err := gqlCtx.RolesGroupsFacade.AddRoleToGroups(options)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(&nonblocking.AdminAPIMutationAddRoleToGroupsExecutedEventPayload{
				AffectedUserIDs: affectedUserIDs,
				RoleID:          roleID,
				GroupIDs:        slice.Map(groups, func(r *model.Group) string { return r.ID }),
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"role": gqlCtx.Roles.Load(roleID),
			}).Value, nil
		},
	},
)

var removeRoleFromGroupsInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RemoveRoleFromGroupsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"roleKey": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the role.",
		},
		"groupKeys": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "The list of group keys.",
		},
	},
})

var removeRoleFromGroupsPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RemoveRoleFromGroupsPayload",
	Fields: graphql.Fields{
		"role": &graphql.Field{
			Type: graphql.NewNonNull(nodeRole),
		},
	},
})

var _ = registerMutationField(
	"removeRoleFromGroups",
	&graphql.Field{
		Description: "Remove the role from the groups.",
		Type:        graphql.NewNonNull(removeRoleFromGroupsPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(removeRoleFromGroupsInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			roleKey := input["roleKey"].(string)

			groupKeyIfaces := input["groupKeys"].([]interface{})
			groupKeys := make([]string, len(groupKeyIfaces))
			for i, v := range groupKeyIfaces {
				groupKeys[i] = v.(string)
			}

			options := &rolesgroups.RemoveRoleFromGroupsOptions{
				RoleKey:   roleKey,
				GroupKeys: groupKeys,
			}

			gqlCtx := GQLContext(p.Context)

			affectedUserIDs, err := gqlCtx.RolesGroupsFacade.ListAllUserIDsByGroupKeys(groupKeys)
			if err != nil {
				return nil, err
			}

			groups, err := gqlCtx.RolesGroupsFacade.ListAllGroupsByKeys(groupKeys)
			if err != nil {
				return nil, err
			}

			roleID, err := gqlCtx.RolesGroupsFacade.RemoveRoleFromGroups(options)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(&nonblocking.AdminAPIMutationRemoveRoleFromGroupsExecutedEventPayload{
				AffectedUserIDs: affectedUserIDs,
				RoleID:          roleID,
				GroupIDs:        slice.Map(groups, func(r *model.Group) string { return r.ID }),
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"role": gqlCtx.Roles.Load(roleID),
			}).Value, nil
		},
	},
)

var addGroupToRolesInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AddGroupToRolesInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"groupKey": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the group.",
		},
		"roleKeys": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "The list of role keys.",
		},
	},
})

var addGroupToRolesPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AddGroupToRolesPayload",
	Fields: graphql.Fields{
		"group": &graphql.Field{
			Type: graphql.NewNonNull(nodeGroup),
		},
	},
})

var _ = registerMutationField(
	"addGroupToRoles",
	&graphql.Field{
		Description: "Add the group to the roles.",
		Type:        graphql.NewNonNull(addGroupToRolesPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(addGroupToRolesInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			groupKey := input["groupKey"].(string)

			roleKeyIfaces := input["roleKeys"].([]interface{})
			roleKeys := make([]string, len(roleKeyIfaces))
			for i, v := range roleKeyIfaces {
				roleKeys[i] = v.(string)
			}

			options := &rolesgroups.AddGroupToRolesOptions{
				GroupKey: groupKey,
				RoleKeys: roleKeys,
			}

			gqlCtx := GQLContext(p.Context)

			affectedUserIDs, err := gqlCtx.RolesGroupsFacade.ListAllUserIDsByGroupKeys([]string{groupKey})
			if err != nil {
				return nil, err
			}

			roles, err := gqlCtx.RolesGroupsFacade.ListAllRolesByKeys(roleKeys)
			if err != nil {
				return nil, err
			}

			groupID, err := gqlCtx.RolesGroupsFacade.AddGroupToRoles(options)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(&nonblocking.AdminAPIMutationAddGroupToRolesExecutedEventPayload{
				AffectedUserIDs: affectedUserIDs,
				GroupID:         groupID,
				RoleIDs:         slice.Map(roles, func(r *model.Role) string { return r.ID }),
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"group": gqlCtx.Groups.Load(groupID),
			}).Value, nil
		},
	},
)

var removeGroupFromRolesInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RemoveGroupFromRolesInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"groupKey": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the group.",
		},
		"roleKeys": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			Description: "The list of role keys.",
		},
	},
})

var removeGroupFromRolesPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RemoveGroupFromRolesPayload",
	Fields: graphql.Fields{
		"group": &graphql.Field{
			Type: graphql.NewNonNull(nodeGroup),
		},
	},
})

var _ = registerMutationField(
	"removeGroupFromRoles",
	&graphql.Field{
		Description: "Remove the group from the roles.",
		Type:        graphql.NewNonNull(removeGroupFromRolesPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(removeGroupFromRolesInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			groupKey := input["groupKey"].(string)

			roleKeyIfaces := input["roleKeys"].([]interface{})
			roleKeys := make([]string, len(roleKeyIfaces))
			for i, v := range roleKeyIfaces {
				roleKeys[i] = v.(string)
			}

			options := &rolesgroups.RemoveGroupFromRolesOptions{
				GroupKey: groupKey,
				RoleKeys: roleKeys,
			}
			gqlCtx := GQLContext(p.Context)

			affectedUserIDs, err := gqlCtx.RolesGroupsFacade.ListAllUserIDsByGroupKeys([]string{groupKey})
			if err != nil {
				return nil, err
			}

			roles, err := gqlCtx.RolesGroupsFacade.ListAllRolesByKeys(roleKeys)
			if err != nil {
				return nil, err
			}

			groupID, err := gqlCtx.RolesGroupsFacade.RemoveGroupFromRoles(options)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(&nonblocking.AdminAPIMutationRemoveGroupFromRolesExecutedEventPayload{
				AffectedUserIDs: affectedUserIDs,
				RoleIDs:         slice.Map(roles, func(r *model.Role) string { return r.ID }),
				GroupID:         groupID,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"group": gqlCtx.Groups.Load(groupID),
			}).Value, nil
		},
	},
)
