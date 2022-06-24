package graphql

import (
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/graphql-go/graphql"
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

var usageType = graphql.NewEnum(graphql.EnumConfig{
	Name: "SubscriptionItemPriceUsageType",
	Values: graphql.EnumValueConfigMap{
		"NONE": &graphql.EnumValueConfig{
			Value: model.UsageTypeNone,
		},
		"SMS": &graphql.EnumValueConfig{
			Value: model.UsageTypeSMS,
		},
		"MAU": &graphql.EnumValueConfig{
			Value: model.UsageTypeMAU,
		},
	},
})

var smsRegion = graphql.NewEnum(graphql.EnumConfig{
	Name: "SubscriptionItemPriceSMSRegion",
	Values: graphql.EnumValueConfigMap{
		"NONE": &graphql.EnumValueConfig{
			Value: model.SMSRegionNone,
		},
		"NORTH_AMERICA": &graphql.EnumValueConfig{
			Value: model.SMSRegionNorthAmerica,
		},
		"OTHER_REGIONS": &graphql.EnumValueConfig{
			Value: model.SMSRegionOtherRegions,
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
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(usageItem))),
		},
	},
})

var usageItem = graphql.NewObject(graphql.ObjectConfig{
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
		"quantity": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
		},
		"currency": &graphql.Field{
			Type: graphql.String,
		},
		"unitAmount": &graphql.Field{
			Type: graphql.Int,
		},
		"totalAmount": &graphql.Field{
			Type: graphql.Int,
		},
	},
})
