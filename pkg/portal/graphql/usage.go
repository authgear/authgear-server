package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

var usage = graphql.NewObject(graphql.ObjectConfig{
	Name: "Usage",
	Fields: graphql.Fields{
		"items": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(usageItem))),
		},
	},
})

var usageType = graphql.NewEnum(graphql.EnumConfig{
	Name: "UsageType",
	Values: graphql.EnumValueConfigMap{
		"NONE": &graphql.EnumValueConfig{
			Value: model.UsageTypeNone,
		},
		"SMS": &graphql.EnumValueConfig{
			Value: model.UsageTypeSMS,
		},
		"WHATSAPP": &graphql.EnumValueConfig{
			Value: model.UsageTypeWhatsapp,
		},
		"MAU": &graphql.EnumValueConfig{
			Value: model.UsageTypeMAU,
		},
	},
})

var smsRegion = graphql.NewEnum(graphql.EnumConfig{
	Name: "UsageSMSRegion",
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

var whatsappRegion = graphql.NewEnum(graphql.EnumConfig{
	Name: "UsageWhatsappRegion",
	Values: graphql.EnumValueConfigMap{
		"NONE": &graphql.EnumValueConfig{
			Value: model.WhatsappRegionNone,
		},
		"NORTH_AMERICA": &graphql.EnumValueConfig{
			Value: model.WhatsappRegionNorthAmerica,
		},
		"OTHER_REGIONS": &graphql.EnumValueConfig{
			Value: model.WhatsappRegionOtherRegions,
		},
	},
})

var usageItem = graphql.NewObject(graphql.ObjectConfig{
	Name: "UsageItem",
	Fields: graphql.Fields{
		"usageType": &graphql.Field{
			Type: graphql.NewNonNull(usageType),
		},
		"smsRegion": &graphql.Field{
			Type: graphql.NewNonNull(smsRegion),
		},
		"whatsappRegion": &graphql.Field{
			Type: graphql.NewNonNull(whatsappRegion),
		},
		"quantity": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
		},
	},
})
