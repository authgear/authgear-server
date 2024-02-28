package graphql

import (
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func init() {
	// Role and group, user and group forms a initialization cycle.
	// So we break the cycle by using AddFieldConfig.
	nodeGroup.AddFieldConfig("roles", &graphql.Field{
		Type:        connRole.ConnectionType,
		Description: "The list of roles this group has.",
		Args:        relay.NewConnectionArgs(graphql.FieldConfigArgument{}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			source := p.Source.(*model.Group)
			gqlCtx := GQLContext(p.Context)

			roles, err := gqlCtx.RolesGroupsFacade.ListRolesByGroupID(source.ID)
			if err != nil {
				return nil, err
			}

			roleIfaces := make([]interface{}, len(roles))
			for i, r := range roles {
				roleIfaces[i] = r
			}

			args := relay.NewConnectionArguments(p.Args)
			return graphqlutil.NewConnectionFromArray(roleIfaces, args), nil
		},
	})

	nodeGroup.AddFieldConfig("users", &graphql.Field{
		Type:        connUser.ConnectionType,
		Description: "The list of users in the group.",
		Args:        relay.NewConnectionArgs(graphql.FieldConfigArgument{}),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			source := p.Source.(*model.Group)
			gqlCtx := GQLContext(p.Context)
			pageArgs := graphqlutil.NewPageArgs(relay.NewConnectionArguments(p.Args))

			refs, result, err := gqlCtx.RolesGroupsFacade.ListUserIDsByGroupID(source.ID, pageArgs)
			if err != nil {
				return nil, err
			}

			var lazyItems []graphqlutil.LazyItem
			for _, ref := range refs {
				lazyItems = append(lazyItems, graphqlutil.LazyItem{
					Lazy:   gqlCtx.Users.Load(ref.ID),
					Cursor: graphqlutil.Cursor(ref.Cursor),
				})
			}

			return graphqlutil.NewConnectionFromResult(lazyItems, result)
		},
	})

}

const typeGroup = "Group"

var nodeGroup = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeGroup,
		Description: "Authgear group",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeGroup),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"key": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The key of the group.",
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional name of the group.",
			},
			"description": &graphql.Field{
				Type:        graphql.String,
				Description: "The optional description of the group.",
			},
		},
	}),
	&model.Group{},
	func(ctx *Context, id string) (interface{}, error) {
		return ctx.Groups.Load(id).Value, nil
	},
)

var connGroup = graphqlutil.NewConnectionDef(nodeGroup)
