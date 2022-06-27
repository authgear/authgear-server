package graphql

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/service"
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
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
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

			// Create the checkout session in stripe
			cs, err := ctx.StripeService.CreateCheckoutSession(appID, user.Email, plan)
			if err != nil {
				return nil, err
			}

			//	Insert subscription checkout record to the db
			_, err = ctx.SubscriptionService.CreateSubscriptionCheckout(cs)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"url": cs.URL,
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
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}
			appID := resolvedNodeID.ID
			ctx := GQLContext(p.Context)
			// Access Control: collaborator.
			_, err := ctx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			// Update checkout session customer id and change the status to completed only
			// Subscription will be created in the webhook
			cs, err := ctx.StripeService.FetchCheckoutSession(checkoutSessionID)
			if err != nil {
				return nil, err
			}
			if !cs.IsCompleted() {
				return nil, apierrors.NewForbidden("the checkout session is not completed")
			}
			if cs.StripeCustomerID == nil {
				return nil, apierrors.NewInvalid("missing customer ID in the completed checkout session")
			}
			err = ctx.SubscriptionService.UpdateSubscriptionCheckoutStatusAndCustomerID(
				appID,
				checkoutSessionID,
				model.SubscriptionCheckoutStatusCompleted,
				*cs.StripeCustomerID,
			)
			if err != nil {
				// The checkout is not found or the checkout is already subscribed
				// Tolerate it.
				if !errors.Is(err, service.ErrSubscriptionCheckoutNotFound) {
					return nil, err
				}
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": ctx.Apps.Load(appID),
			}).Value, nil

		},
	},
)

var generateStripeCustomerPortalSessionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "GenerateStripeCustomerPortalSessionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target app ID.",
		},
	},
})

var generateStripeCustomerPortalSessionPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "GenerateStripeCustomerPortalSessionPayload",
	Fields: graphql.Fields{
		"url": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

var _ = registerMutationField(
	"generateStripeCustomerPortalSession",
	&graphql.Field{
		Description: "Generate Stripe customer portal session",
		Type:        graphql.NewNonNull(generateStripeCustomerPortalSessionPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(generateStripeCustomerPortalSessionInput),
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

			resolvedNodeID := relay.FromGlobalID(appNodeID)
			if resolvedNodeID == nil || resolvedNodeID.Type != typeApp {
				return nil, apierrors.NewInvalid("invalid app ID")
			}

			appID := resolvedNodeID.ID
			ctx := GQLContext(p.Context)

			// Access Control: collaborator.
			_, err := ctx.AuthzService.CheckAccessOfViewer(appID)
			if err != nil {
				return nil, err
			}

			sub, err := ctx.SubscriptionService.GetSubscription(appID)
			if err != nil {
				return nil, err
			}

			s, err := ctx.StripeService.GenerateCustomerPortalSession(appID, sub.StripeCustomerID)
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"url": s.URL,
			}).Value, nil
		},
	},
)
