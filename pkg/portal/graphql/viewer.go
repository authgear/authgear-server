package graphql

import (
	"context"

	"github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
)

const typeViewer = "Viewer"

var nodeViewer = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name:        typeViewer,
		Description: "The viewer",
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": relay.GlobalIDField(typeViewer, nil),
			"email": &graphql.Field{
				Type: graphql.String,
			},
			"projectQuota": &graphql.Field{
				Type: graphql.Int,
			},
			"projectOwnerCount": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
	}),
	&model.User{},
	func(ctx context.Context, id string) (interface{}, error) {
		gqlCtx := GQLContext(ctx)

		// Ensure only the authenticated user can fetch their own viewer.
		sessionInfo := session.GetValidSessionInfo(ctx)
		if sessionInfo == nil {
			return nil, nil
		}
		if sessionInfo.UserID != id {
			return nil, nil
		}

		return gqlCtx.Users.Load(id).Value, nil
	},
)
