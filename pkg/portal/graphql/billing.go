package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

var priceType = graphql.NewEnum(graphql.EnumConfig{
	Name: "SubscriptionItemPriceType",
	Values: graphql.EnumValueConfigMap{
		"FIXED": &graphql.EnumValueConfig{
			Value: model.PriceTypeFixed,
		},
		"USAGE": &graphql.EnumValueConfig{
			Value: model.PriceTypeUsage,
		},
	},
})

var transformQuantityRound = graphql.NewEnum(graphql.EnumConfig{
	Name: "TransformQuantityRound",
	Values: graphql.EnumValueConfigMap{
		"NONE": &graphql.EnumValueConfig{
			Value: model.TransformQuantityRoundNone,
		},
		"UP": &graphql.EnumValueConfig{
			Value: model.TransformQuantityRoundUp,
		},
		"DOWN": &graphql.EnumValueConfig{
			Value: model.TransformQuantityRoundDown,
		},
	},
})

var price = graphql.NewObject(graphql.ObjectConfig{
	Name: "SubscriptionItemPrice",
	Fields: graphql.Fields{
		"currency": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"unitAmount": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"type": &graphql.Field{
			Type: graphql.NewNonNull(priceType),
		},
		"usageType": &graphql.Field{
			Type: graphql.NewNonNull(usageType),
		},
		"smsRegion": &graphql.Field{
			Type: graphql.NewNonNull(smsRegion),
		},
		"whatsappRegion": &graphql.Field{
			Type: graphql.NewNonNull(whatsappRegion),
		},
		"transformQuantityDivideBy": &graphql.Field{
			Type: graphql.Int,
		},
		"transformQuantityRound": &graphql.Field{
			Type: graphql.NewNonNull(transformQuantityRound),
		},
		"freeQuantity": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var subscriptionPlan = graphql.NewObject(graphql.ObjectConfig{
	Name: "SubscriptionPlan",
	Fields: graphql.Fields{
		"name": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
		},
		"prices": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(price))),
		},
	},
})

var subscriptionUsage = graphql.NewObject(graphql.ObjectConfig{
	Name: "SubscriptionUsage",
	Fields: graphql.Fields{
		"nextBillingDate": &graphql.Field{
			Type: graphql.NewNonNull(graphql.DateTime),
		},
		"items": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(subscriptionUsageItem))),
		},
	},
})

var subscription = graphql.NewObject(graphql.ObjectConfig{
	Name: "Subscription",
	Fields: graphql.Fields{
		"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"createdAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
		"updatedAt":   &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
		"cancelledAt": &graphql.Field{Type: graphql.DateTime},
		"endedAt":     &graphql.Field{Type: graphql.DateTime},
	},
})

var subscriptionUsageItem = graphql.NewObject(graphql.ObjectConfig{
	Name: "SubscriptionUsageItem",
	Fields: graphql.Fields{
		"type": &graphql.Field{
			Type: graphql.NewNonNull(priceType),
		},
		"usageType": &graphql.Field{
			Type: graphql.NewNonNull(usageType),
		},
		"smsRegion": &graphql.Field{
			Type: graphql.NewNonNull(smsRegion),
		},
		"whatsappRegion": &graphql.Field{
			Type: graphql.NewNonNull(whatsappRegion),
		},
		"currency": &graphql.Field{
			Type: graphql.String,
		},
		"unitAmount": &graphql.Field{
			Type: graphql.Int,
		},
		"transformQuantityDivideBy": &graphql.Field{
			Type: graphql.Int,
		},
		"transformQuantityRound": &graphql.Field{
			Type: graphql.NewNonNull(transformQuantityRound),
		},
		"freeQuantity": &graphql.Field{
			Type: graphql.Int,
		},

		"quantity": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"totalAmount": &graphql.Field{
			Type: graphql.Int,
		},
	},
})
