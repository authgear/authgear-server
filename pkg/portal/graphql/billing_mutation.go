package graphql

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"
)

var createCheckoutSessionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateCheckoutSessionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "App ID.",
		},
		"planName": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Plan name.",
		},
	},
})

var createCheckoutSessionPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateCheckoutSessionPayload",
	Fields: graphql.Fields{
		"url": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

var _ = registerMutationField(
	"createCheckoutSession",
	&graphql.Field{
		Description: "Create stripe checkout session",
		Type:        graphql.NewNonNull(createCheckoutSessionPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(createCheckoutSessionInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenticated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, AccessDenied.New("only authenticated users can subscribe to a plan")
			}

			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			planName := input["planName"].(string)

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID
			gqlCtx := GQLContext(p.Context)
			// Access Control: collaborator.
			_, err := gqlCtx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			// fetch the subscription plan
			ctx := GQLContext(p.Context)
			plan, err := ctx.StripeService.GetSubscriptionPlan(planName)
			if err != nil {
				return nil, apierrors.NewInvalid("invalid plan name")
			}

			// fetch the current user email
			val, err := ctx.Users.Load(sessionInfo.UserID).Value()
			if err != nil {
				return nil, apierrors.NewInvalid("failed to load current user")
			}
			user, ok := val.(*model.User)
			if !ok {
				return nil, apierrors.NewInvalid("failed to load current user")
			}

			// create the checkout session
			url, err := ctx.StripeService.CreateCheckoutSession(appID, user.Email, plan)
			if err != nil {
				return nil, err
			}
			return graphqlutil.NewLazyValue(map[string]interface{}{
				"url": url,
			}).Value, nil
		},
	},
)

var reconcileCheckoutSessionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "reconcileCheckoutSession",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target app ID.",
		},
		"checkoutSessionID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "Checkout session ID.",
		},
	},
})

var reconcileCheckoutSessionPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "reconcileCheckoutSessionPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"reconcileCheckoutSession",
	&graphql.Field{
		Description: "Reconcile the completed checkout session",
		Type:        graphql.NewNonNull(reconcileCheckoutSessionPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(reconcileCheckoutSessionInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenticated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, AccessDenied.New("only authenticated users can create domain")
			}

			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			checkoutSessionID := input["checkoutSessionID"].(string)

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID
			ctx := GQLContext(p.Context)
			// Access Control: collaborator.
			_, err := ctx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			// create the stripe subscription
			sub, err := ctx.StripeService.CreateSubscription(checkoutSessionID)
			if err != nil {
				return nil, err
			}
			if appID != sub.AppID {
				return nil, fmt.Errorf("mismatched app id")
			}

			_, err = ctx.SubscriptionService.CreateSubscription(sub)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": ctx.Apps.Load(appID),
			}).Value, nil

		},
	},
)
