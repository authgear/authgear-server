package graphql

import (
	"errors"

	relay "github.com/authgear/graphql-go-relay"
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
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
				return nil, Unauthenticated.New("only authenticated users can subscribe to a plan")
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

			// Insert subscription checkout record to the db
			checkout, err := ctx.SubscriptionService.CreateSubscriptionCheckout(cs)
			if err != nil {
				return nil, err
			}

			app, err := gqlCtx.AppService.Get(appID)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuditService.Log(app, &nonblocking.ProjectBillingCheckoutCreatedEventPayload{
				SubscriptionCheckoutID: checkout.ID,
				PlanName:               planName,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"url": cs.URL,
			}).Value, nil
		},
	},
)

var previewUpdateSubscriptionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "PreviewUpdateSubscriptionInput",
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

var previewUpdateSubscriptionPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "PreviewUpdateSubscriptionPayload",
	Fields: graphql.Fields{
		"currency":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"amountDue": &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
	},
})

var _ = registerMutationField(
	"previewUpdateSubscription",
	&graphql.Field{
		Description: "Preview update subscription",
		Type:        graphql.NewNonNull(previewUpdateSubscriptionPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(previewUpdateSubscriptionInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenticated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can preview update subscription")
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

			// Fetch the subscription plan
			ctx := GQLContext(p.Context)
			plan, err := ctx.StripeService.GetSubscriptionPlan(planName)
			if err != nil {
				return nil, err
			}

			// Fetch the subscription
			subscription, err := ctx.SubscriptionService.GetSubscription(appID)
			if err != nil {
				return nil, err
			}

			// Fetch the app
			app, err := ctx.AppService.Get(appID)
			if err != nil {
				return nil, err
			}

			// Changing to the same plan is disallowed.
			if app.Context.PlanName == planName {
				return nil, apierrors.NewInvalid("changing to the same plan is disallowed")
			}

			preview, err := ctx.StripeService.PreviewUpdateSubscription(
				subscription.StripeSubscriptionID,
				plan,
			)
			if err != nil {
				return nil, err
			}

			return preview, nil
		},
	},
)

var updateSubscriptionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "UpdateSubscriptionInput",
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

var updateSubscriptionPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "UpdateSubscriptionPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"updateSubscription",
	&graphql.Field{
		Description: "Update subscription",
		Type:        graphql.NewNonNull(updateSubscriptionPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(updateSubscriptionInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenticated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can update subscription")
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

			// Fetch the subscription plan
			ctx := GQLContext(p.Context)
			plan, err := ctx.StripeService.GetSubscriptionPlan(planName)
			if err != nil {
				return nil, err
			}

			// Fetch the subscription
			subscription, err := ctx.SubscriptionService.GetSubscription(appID)
			if err != nil {
				return nil, err
			}

			// Fetch the app
			app, err := ctx.AppService.Get(appID)
			if err != nil {
				return nil, err
			}

			// Changing to the same plan is disallowed.
			if app.Context.PlanName == planName {
				return nil, apierrors.NewInvalid("changing to the same plan is disallowed")
			}

			err = ctx.StripeService.UpdateSubscription(
				subscription.StripeSubscriptionID,
				plan,
			)
			if err != nil {
				return nil, err
			}

			err = ctx.SubscriptionService.UpdateAppPlan(
				appID,
				planName,
			)
			if err != nil {
				return nil, err
			}

			err = gqlCtx.AuditService.Log(app, &nonblocking.ProjectBillingSubscriptionUpdatedEventPayload{
				SubscriptionID: subscription.ID,
				PlanName:       planName,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": ctx.Apps.Load(appID),
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
				return nil, Unauthenticated.New("only authenticated users can create domain")
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
			err = ctx.SubscriptionService.MarkCheckoutCompleted(
				appID,
				checkoutSessionID,
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
				return nil, Unauthenticated.New("only authenticated users can create domain")
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

var setSubscriptionCancelledStatusInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "SetSubscriptionCancelledStatusInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target app ID.",
		},
		"cancelled": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Target app subscription cancellation status.",
		},
	},
})

var setSubscriptionCancelledStatusPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "SetSubscriptionCancelledStatusPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"setSubscriptionCancelledStatus",
	&graphql.Field{
		Description: "Set app subscription cancellation status",
		Type:        graphql.NewNonNull(setSubscriptionCancelledStatusPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(setSubscriptionCancelledStatusInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenticated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can set subscription cancelled status")
			}

			input := p.Args["input"].(map[string]interface{})
			appNodeID := input["appID"].(string)
			cancelled, _ := input["cancelled"].(bool)
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

			subscription, err := ctx.SubscriptionService.GetSubscription(appID)
			if err != nil {
				return nil, err
			}

			periodEnd, err := ctx.StripeService.SetSubscriptionCancelAtPeriodEnd(subscription.StripeSubscriptionID, cancelled)
			if err != nil {
				return nil, err
			}

			err = ctx.SubscriptionService.SetSubscriptionCancelledStatus(subscription.ID, cancelled, periodEnd)
			if err != nil {
				return nil, err
			}

			app, err := ctx.AppService.Get(appID)
			if err != nil {
				return nil, err
			}

			err = ctx.AuditService.Log(app, &nonblocking.ProjectBillingSubscriptionStatusUpdatedEventPayload{
				SubscriptionID: subscription.ID,
				Cancelled:      cancelled,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": ctx.Apps.Load(appID),
			}).Value, nil
		},
	},
)

var cancelFailedSubscriptionInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "cancelFailedSubscriptionInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"appID": &graphql.InputObjectFieldConfig{
			Type:        graphql.NewNonNull(graphql.ID),
			Description: "Target app ID.",
		},
	},
})

var cancelFailedSubscriptionPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CancelFailedSubscriptionPayload",
	Fields: graphql.Fields{
		"app": &graphql.Field{Type: graphql.NewNonNull(nodeApp)},
	},
})

var _ = registerMutationField(
	"cancelFailedSubscription",
	&graphql.Field{
		Description: "Cancel failed subscription",
		Type:        graphql.NewNonNull(cancelFailedSubscriptionPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(cancelFailedSubscriptionInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			// Access Control: authenticated user.
			sessionInfo := session.GetValidSessionInfo(p.Context)
			if sessionInfo == nil {
				return nil, Unauthenticated.New("only authenticated users can cancel failed subscription")
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

			customerID, err := ctx.SubscriptionService.GetLastProcessingCustomerID(appID)
			if err != nil {
				return nil, err
			}
			if customerID == nil {
				return nil, apierrors.NewInvalid("last completed checkout session not found")
			}

			subscription, err := ctx.StripeService.GetSubscription(*customerID)
			if err != nil {
				return nil, err
			}

			if subscription == nil ||
				subscription.LatestInvoice == nil ||
				subscription.LatestInvoice.PaymentIntent == nil ||
				subscription.LatestInvoice.PaymentIntent.LastPaymentError == nil {
				// only allow cancelling failed subscription
				// normal subscription should be cancelled via setSubscriptionCancelledStatus
				return nil, apierrors.NewInvalid("subscription not found or the subscription doesn't have payment error")
			}

			err = ctx.StripeService.CancelSubscriptionImmediately(subscription.ID)
			if err != nil {
				ctx.Logger().WithError(err).Error("failed to cancel subscription")
				return nil, apierrors.NewInternalError("failed to cancel subscription")
			}

			// After cancelling failed subscription
			// webhook event `customer.subscription.updated` will be fired
			// and the subscription will change to `incomplete_expired` status
			//
			// although the status will be changed by webhook
			// we set it to expiry first to avoid ui inconsistent before the webhook come
			err = ctx.SubscriptionService.MarkCheckoutExpired(appID, *customerID)
			if err != nil {
				ctx.Logger().WithError(err).Error("failed to update checkout session status")
				return nil, apierrors.NewInternalError("failed to update checkout session status")
			}

			app, err := ctx.AppService.Get(appID)
			if err != nil {
				return nil, err
			}

			err = ctx.AuditService.Log(app, &nonblocking.ProjectBillingSubscriptionCancelledEventPayload{
				SubscriptionID: subscription.ID,
				CustomerID:     *customerID,
			})
			if err != nil {
				return nil, err
			}

			return graphqlutil.NewLazyValue(map[string]interface{}{
				"app": ctx.Apps.Load(appID),
			}).Value, nil
		},
	},
)
