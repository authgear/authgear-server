package graphql

import (
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func init() {
	// Role and group forms a initialization cycle.
	// So we break the cycle by using AddFieldConfig.
	nodeRole.AddFieldConfig("groups", &graphql.Field{
		Type:        connGroup.ConnectionType,
		Description: "The list of groups this role is in.",
		Args:        relay.NewConnectionArgs(graphql.FieldConfigArgument{}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			source := p.Source.(*model.Role)
			gqlCtx := GQLContext(p.Context)

			groups, err := gqlCtx.RolesGroupsFacade.ListGroupsByRoleID(source.ID)
			if err != nil {
				return nil, err
			}

			groupIfaces := make([]interface{}, len(groups))
			for i, g := range groups {
				groupIfaces[i] = g
			}

			args := relay.NewConnectionArguments(p.Args)
			return graphqlutil.NewConnectionFromArray(groupIfaces, args), nil
		},
	})
}

const typeRole = "Role"

var nodeRole = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeRole,
		Description: "Authgear role",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeRole),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"key": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The key of the role.",
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional name of the role.",
			},
			"description": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional description of the role.",
			},
		},
	}),
	&model.Role{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Roles.Load(id).Value, nil
	},
)

var connRole = graphqlutil.NewConnectionDef(nodeRole)
