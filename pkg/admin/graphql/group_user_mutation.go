package graphql

import (
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/graphql-go/graphql"
)

var addGroupToUsersInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "AddGroupToUsersInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"groupKey": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the group.",
		},
		"userIDs": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.ID)),
			Description: "The list of user ids.",
		},
	},
})

var addGroupToUsersPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "AddGroupToUsersPayload",
	Fields: graphql.Fields{
		"group": &graphql.Field{
			Type: graphql.NewNonNull(nodeGroup),
		},
	},
})

var _ = registerMutationField(
	"addGroupToUsers",
	&graphql.Field{
		Description: "Add the group to the users.",
		Type:        graphql.NewNonNull(addGroupToUsersPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(addGroupToUsersInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			groupKey := input["groupKey"].(string)
			userIDIfaces := input["userIDs"].([]interface{})
			userIDs := make([]string, len(userIDIfaces))
			for i, v := range userIDIfaces {
				userIDs[i] = v.(string)
			}
			gqlCtx := GQLContext(p.Context)

			options := &rolesgroups.AddGroupToUsersOptions{
				GroupKey: groupKey,
				UserIDs:  userIDs,
			}
			groupID, err := gqlCtx.RolesGroupsFacade.AddGroupToUsers(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"group": gqlCtx.Groups.Load(groupID),
			}).Value, nil

		},
	},
)

var removeGroupFromUsersInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RemoveGroupFromUsersInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"groupKey": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The key of the group.",
		},
		"userIDs": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewList(graphql.NewNonNull(graphql.ID)),
			Description: "The list of user ids.",
		},
	},
})

var removeGroupFromUsersPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "RemoveGroupToUsersPayload",
	Fields: graphql.Fields{
		"group": &graphql.Field{
			Type: graphql.NewNonNull(nodeGroup),
		},
	},
})

var _ = registerMutationField(
	"removeGroupFromUsers",
	&graphql.Field{
		Description: "Remove the group to the users.",
		Type:        graphql.NewNonNull(removeGroupFromUsersPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(removeGroupFromUsersInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(map[string]interface{})

			groupKey := input["groupKey"].(string)
			userIDIfaces := input["userIDs"].([]interface{})
			userIDs := make([]string, len(userIDIfaces))
			for i, v := range userIDIfaces {
				userIDs[i] = v.(string)
			}
			gqlCtx := GQLContext(p.Context)

			options := &rolesgroups.RemoveGroupFromUsersOptions{
				GroupKey: groupKey,
				UserIDs:  userIDs,
			}
			groupID, err := gqlCtx.RolesGroupsFacade.RemoveGroupFromUsers(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"group": gqlCtx.Groups.Load(groupID),
			}).Value, nil

		},
	},
)
