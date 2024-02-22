package graphql

import (
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
