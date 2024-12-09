package graphql

import (
	"errors"
	"fmt"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

var checkCollaboratorInvitationPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CheckCollaboratorInvitationPayload",
	Fields: graphql.Fields{
		"isInvitee": &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		"appID":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

var query = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"node":  nodeDefs.NodeField,
		"nodes": nodeDefs.NodesField,
		"viewer": &graphql.Field{
			Description: "The current viewer",
			Type:        nodeViewer,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				sessionInfo := session.GetValidSessionInfo(ctx)
				if sessionInfo == nil {
					return nil, nil
				}

				return viewerSubresolver(ctx, gqlCtx, sessionInfo.UserID)
			},
		},
		"appList": &graphql.Field{
			Description: "The list of apps accessible to the current viewer",
			Type: graphql.NewList(graphql.NewNonNull(graphql.NewObject(graphql.ObjectConfig{
				Name: "AppListItem",
				Fields: graphql.Fields{
					"appID":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
					"publicOrigin": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
				},
			}))),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				sessionInfo := session.GetValidSessionInfo(ctx)
				if sessionInfo == nil {
					return nil, nil
				}

				// Access control is not needed here cause List returns accessible apps.
				apps, err := gqlCtx.AppService.GetAppList(ctx, sessionInfo.UserID)
				if err != nil {
					return nil, err
				}

				return apps, nil
			},
		},
		"checkCollaboratorInvitation": &graphql.Field{
			Description: "Check whether the viewer can accept the collaboration invitation",
			Type:        checkCollaboratorInvitationPayload,
			Args:        graphql.FieldConfigArgument{"code": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)}},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)

				code := p.Args["code"].(string)

				invitation, err := gqlCtx.CollaboratorService.GetInvitationWithCode(ctx, code)
				if err != nil {
					if errors.Is(err, service.ErrCollaboratorInvitationInvalidCode) {
						return nil, nil
					}
					return nil, err
				}

				sessionInfo := session.GetValidSessionInfo(ctx)
				if sessionInfo == nil {
					return graphqlutil.NewLazyValue(map[string]interface{}{
						"isInvitee": false,
						"appID":     invitation.AppID,
					}).Value, nil
				}
				actorID := sessionInfo.UserID

				err = gqlCtx.CollaboratorService.CheckInviteeEmail(ctx, invitation, actorID)
				if err != nil {
					if errors.Is(err, service.ErrCollaboratorInvitationInvalidEmail) {
						return graphqlutil.NewLazyValue(map[string]interface{}{
							"isInvitee": false,
							"appID":     invitation.AppID,
						}).Value, nil
					}
					return nil, err
				}

				return graphqlutil.NewLazyValue(map[string]interface{}{
					"isInvitee": true,
					"appID":     invitation.AppID,
				}).Value, nil
			},
		},
		"activeUserChart": &graphql.Field{
			Description: "Active users chart dataset",
			Type:        activeUserChart,
			Args: newAnalyticArgs(graphql.FieldConfigArgument{
				"periodical": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(periodicalEnum),
				},
			}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context

				gqlCtx := GQLContext(ctx)
				periodical := p.Args["periodical"].(string)
				appID, rangeFrom, rangeTo, err := getAnalyticArgs(p.Args)
				if err != nil {
					return nil, err
				}

				// Access Control: collaborator.
				_, err = gqlCtx.AuthzService.CheckAccessOfViewer(ctx, appID)
				if err != nil {
					return nil, nil
				}

				err = checkChartDateRangeInput(rangeFrom, rangeTo)
				if err != nil {
					return nil, err
				}

				chart, err := gqlCtx.AnalyticChartService.GetActiveUserChart(
					ctx,
					appID,
					periodical,
					*rangeFrom,
					*rangeTo,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch dataset: %w", err)
				}
				return graphqlutil.NewLazyValue(chart).Value, nil
			},
		},
		"totalUserCountChart": &graphql.Field{
			Description: "Total users count chart dataset",
			Type:        totalUserCountChart,
			Args:        newAnalyticArgs(graphql.FieldConfigArgument{}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context

				gqlCtx := GQLContext(ctx)
				appID, rangeFrom, rangeTo, err := getAnalyticArgs(p.Args)
				if err != nil {
					return nil, err
				}

				// Access Control: collaborator.
				_, err = gqlCtx.AuthzService.CheckAccessOfViewer(ctx, appID)
				if err != nil {
					return nil, nil
				}

				err = checkChartDateRangeInput(rangeFrom, rangeTo)
				if err != nil {
					return nil, err
				}

				chart, err := gqlCtx.AnalyticChartService.GetTotalUserCountChart(
					ctx,
					appID,
					*rangeFrom,
					*rangeTo,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch dataset: %w", err)
				}
				return graphqlutil.NewLazyValue(chart).Value, nil
			},
		},
		"signupConversionRate": &graphql.Field{
			Description: "Signup conversion rate dashboard data",
			Type:        signupConversionRate,
			Args:        newAnalyticArgs(graphql.FieldConfigArgument{}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)
				appID, rangeFrom, rangeTo, err := getAnalyticArgs(p.Args)
				if err != nil {
					return nil, err
				}

				// Access Control: collaborator.
				_, err = gqlCtx.AuthzService.CheckAccessOfViewer(ctx, appID)
				if err != nil {
					return nil, nil
				}

				err = checkChartDateRangeInput(rangeFrom, rangeTo)
				if err != nil {
					return nil, err
				}

				signupConversionRateData, err := gqlCtx.AnalyticChartService.GetSignupConversionRate(
					ctx,
					appID,
					*rangeFrom,
					*rangeTo,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch conversion rate data: %w", err)
				}
				return graphqlutil.NewLazyValue(signupConversionRateData).Value, nil
			},
		},
		"signupByMethodsChart": &graphql.Field{
			Description: "Signup by methods dataset",
			Type:        signupByMethodsChart,
			Args:        newAnalyticArgs(graphql.FieldConfigArgument{}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)
				appID, rangeFrom, rangeTo, err := getAnalyticArgs(p.Args)
				if err != nil {
					return nil, err
				}

				// Access Control: collaborator.
				_, err = gqlCtx.AuthzService.CheckAccessOfViewer(ctx, appID)
				if err != nil {
					return nil, nil
				}

				err = checkChartDateRangeInput(rangeFrom, rangeTo)
				if err != nil {
					return nil, err
				}

				chart, err := gqlCtx.AnalyticChartService.GetSignupByMethodsChart(
					ctx,
					appID,
					*rangeFrom,
					*rangeTo,
				)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch dataset: %w", err)
				}
				return graphqlutil.NewLazyValue(chart).Value, nil
			},
		},
		"subscriptionPlans": &graphql.Field{
			Description: "Available subscription plans",
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(subscriptionPlan))),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)
				plans, err := gqlCtx.StripeService.FetchSubscriptionPlans(ctx)
				if err != nil {
					return nil, err
				}
				return plans, nil
			},
		},
	},
})
