package graphql

import (
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/graphql-go/graphql"
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
			roleID, err := gqlCtx.RolesGroupsFacade.AddRoleToGroups(options)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"role": gqlCtx.Roles.Load(roleID),
			}).Value, nil
		},
	},
)
