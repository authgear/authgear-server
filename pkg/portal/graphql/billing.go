package graphql

import (
	"github.com/authgear/authgear-server/pkg/portal/libstripe"
	"github.com/graphql-go/graphql"
)

var priceType = graphql.NewEnum(graphql.EnumConfig{
	Name: "SubscriptionItemPriceType",
	Values: graphql.EnumValueConfigMap{
		"FIXED": &graphql.EnumValueConfig{
			Value: libstripe.PriceTypeFixed,
		},
		"USAGE": &graphql.EnumValueConfig{
			Value: libstripe.PriceTypeUsage,
		},
	},
})

var usageType = graphql.NewEnum(graphql.EnumConfig{
	Name: "SubscriptionItemPriceUsageType",
	Values: graphql.EnumValueConfigMap{
		"NONE": &graphql.EnumValueConfig{
			Value: libstripe.UsageTypeNone,
		},
		"SMS": &graphql.EnumValueConfig{
			Value: libstripe.UsageTypeSMS,
		},
	},
})

var smsRegion = graphql.NewEnum(graphql.EnumConfig{
	Name: "SubscriptionItemPriceSMSRegion",
	Values: graphql.EnumValueConfigMap{
		"NONE": &graphql.EnumValueConfig{
			Value: libstripe.SMSRegionNone,
		},
		"NORTH_AMERICA": &graphql.EnumValueConfig{
			Value: libstripe.SMSRegionNorthAmerica,
		},
		"OTHER_REGIONS": &graphql.EnumValueConfig{
			Value: libstripe.SMSRegionOtherRegions,
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
			Type: graphql.NewNonNull(graphql.NewList(price)),
		},
	},
})
