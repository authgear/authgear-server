package graphql

import (
	"context"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/geoip"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

const typeViewer = "Viewer"

var viewerSubresolver = func(gqlCtx *Context, id string) (interface{}, error) {
	userIface, err := gqlCtx.Users.Load(id).Value()
	if err != nil {
		return nil, err
	}

	user := userIface.(*model.User)

	requestIP := httputil.GetIP(gqlCtx.Request, bool(gqlCtx.TrustProxy))
	geoipInfo, ok := geoip.DefaultDatabase.IPString(requestIP)
	if ok {
		user.GeoIPCountryCode = geoipInfo.CountryCode
	}

	isCompleted, err := gqlCtx.OnboardService.CheckCompletion(id)
	if err != nil {
		return nil, err
	}
	user.IsOnboardingSurveyCompleted = isCompleted

	return user, nil
}

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
			"formattedName": &graphql.Field{
				Type: graphql.String,
			},
			"projectQuota": &graphql.Field{
				Type: graphql.Int,
			},
			"projectOwnerCount": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"geoIPCountryCode": &graphql.Field{
				Type: graphql.String,
			},
			"isOnboardingSurveyCompleted": &graphql.Field{
				Type: graphql.Boolean,
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

		return viewerSubresolver(gqlCtx, id)
	},
)
