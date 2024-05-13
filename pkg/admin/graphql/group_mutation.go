package graphql

import (
	relay "github.com/authgear/graphql-go-relay"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

var createGroupInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateGroupInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"key": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the group.",
		},
		"name": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The optional name of the group.",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The optional description of the group.",
		},
	},
})

var createGroupPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateGroupPayload",
	Fields: graphql.Fields{
		"group": &graphql.Field{
			Type: graphql.NewNonNull(nodeGroup),
		},
	},
})

var _ = registerMutationField(
	"createGroup",
	&graphql.Field{
		Description: "Create a new group.",
		Type:        graphql.NewNonNull(createGroupPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createGroupInput),
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

			options := &rolesgroups.NewGroupOptions{
				Key:         key,
				Name:        name,
				Description: description,
			}

			gqlCtx := GQLContext(p.Context)
			groupID, err := gqlCtx.RolesGroupsFacade.CreateGroup(options)
			if err != nil {
				return nil, err
			}

			group, err := gqlCtx.RolesGroupsFacade.GetGroup(groupID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(&nonblocking.AdminAPIMutationCreateGroupExecutedEventPayload{
				Group: *group,
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

var updateGroupInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateGroupInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the group.",
		},
		"key": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The new key of the group. Pass null if you do not need to update the key.",
		},
		"name": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The new name of the group. Pass null if you do not need to update the name. Pass an empty string to remove the name.",
		},
		"description": &graphql.InputObjectFieldConfig{
			Type:        graphql.String,
			Description: "The new description of the group. Pass null if you do not need to update the description. Pass an empty string to remove the description.",
		},
	},
})

var updateGroupPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateGroupPayload",
	Fields: graphql.Fields{
		"group": &graphql.Field{
			Type: graphql.NewNonNull(nodeGroup),
		},
	},
})

var _ = registerMutationField(
	"updateGroup",
	&graphql.Field{
		Description: "Update an existing group.",
		Type:        graphql.NewNonNull(updateGroupPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateGroupInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			groupNodeID := input["id"].(string)

			resolvedNodeID := relay.FromGlobalID(groupNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeGroup {
				return nil, apierrors.NewInvalid("invalid group ID")
			}
			groupID := resolvedNodeID.ID

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

			options := &rolesgroups.UpdateGroupOptions{
				ID:             groupID,
				NewKey:         newKey,
				NewName:        newName,
				NewDescription: newDescription,
			}

			gqlCtx := GQLContext(p.Context)

			originalGroup, err := gqlCtx.RolesGroupsFacade.GetGroup(groupID)
			if err != nil {
				return nil, err
			}

			affectedUserIDs, err := gqlCtx.RolesGroupsFacade.ListAllUserIDsByGroupIDs([]string{groupID})
			if err != nil {
				return nil, err
			}

			err = gqlCtx.RolesGroupsFacade.UpdateGroup(options)
			if err != nil {
				return nil, err
			}

			newGroup, err := gqlCtx.RolesGroupsFacade.GetGroup(groupID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(&nonblocking.AdminAPIMutationUpdateGroupExecutedEventPayload{
				AffectedUserIDs: affectedUserIDs,
				OriginalGroup:   *originalGroup,
				NewGroup:        *newGroup,
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

var deleteGroupInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "DeleteGroupInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "The ID of the group.",
		},
	},
})

var deleteGroupPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeleteGroupPayload",
	Fields: graphql.Fields{
		"ok": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

var _ = registerMutationField(
	"deleteGroup",
	&graphql.Field{
		Description: "Delete an existing group. The associations between the group with other roles and other users will also be deleted.",
		Type:        graphql.NewNonNull(deleteGroupPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(deleteGroupInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			groupNodeID := input["id"].(string)

			resolvedNodeID := relay.FromGlobalID(groupNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeGroup {
				return nil, apierrors.NewInvalid("invalid group ID")
			}
			groupID := resolvedNodeID.ID

			gqlCtx := GQLContext(p.Context)

			group, err := gqlCtx.RolesGroupsFacade.GetGroup(groupID)
			if err != nil {
				return nil, err
			}

			groupRoles, err := gqlCtx.RolesGroupsFacade.ListRolesByGroupID(groupID)
			if err != nil {
				return nil, err
			}

			groupUserIds, err := gqlCtx.RolesGroupsFacade.ListAllUserIDsByGroupIDs([]string{groupID})
			if err != nil {
				return nil, err
			}

			err = gqlCtx.RolesGroupsFacade.DeleteGroup(groupID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.Events.DispatchEventOnCommit(&nonblocking.AdminAPIMutationDeleteGroupExecutedEventPayload{
				Group:        *group,
				GroupRoleIDs: slice.Map(groupRoles, func(r *model.Role) string { return r.ID }),
				GroupUserIDs: groupUserIds,
			})
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"ok": true,
			}, nil
		},
	},
)
