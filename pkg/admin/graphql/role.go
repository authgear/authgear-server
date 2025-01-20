package graphql

import (
	"context"

	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func init() {
	// Role and user, role and group forms a initialization cycle.
	// So we break the cycle by using AddFieldConfig.
	nodeRole.AddFieldConfig("groups", &graphql.Field{
		Type:        connGroup.ConnectionType,
		Description: "The list of groups this role is in.",
		Args:        relay.NewConnectionArgs(graphql.FieldConfigArgument{}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			source := p.Source.(*model.Role)
			ctx := p.Context
			gqlCtx := GQLContext(ctx)

			groups, err := gqlCtx.RolesGroupsFacade.ListGroupsByRoleID(ctx, source.ID)
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

	nodeRole.AddFieldConfig("users", &graphql.Field{
		Type:        connUser.ConnectionType,
		Description: "The list of users who has this role.",
		Args:        relay.NewConnectionArgs(graphql.FieldConfigArgument{}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			source := p.Source.(*model.Role)
			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			pageArgs := graphqlutil.NewPageArgs(relay.NewConnectionArguments(p.Args))

			refs, result, err := gqlCtx.RolesGroupsFacade.ListUserIDsByRoleID(ctx, source.ID, pageArgs)
			if err != nil {
				return nil, err
			}

			var lazyItems []graphqlutil.LazyItem
			for _, ref := range refs {
				lazyItems = append(lazyItems, graphqlutil.LazyItem{
					Lazy:   gqlCtx.Users.Load(ctx, ref.ID),
					Cursor: graphqlutil.Cursor(ref.Cursor),
				})
			}

			return graphqlutil.NewConnectionFromResult(lazyItems, result)
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
	func(ctx context.Context, gqlCtx *Context, id string) (interface{}, error) {
		return gqlCtx.Roles.Load(ctx, id).Value, nil
	},
)

var connRole = graphqlutil.NewConnectionDef(nodeRole)
