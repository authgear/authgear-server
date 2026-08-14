package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/lib/audit"
)

var fraudProtectionOverviewTimeBucketType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FraudProtectionOverviewTimeBucket",
	Fields: graphql.Fields{
		"hour": &graphql.Field{
			Type: graphql.NewNonNull(graphql.DateTime),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewTimeBucket)
				return source.Hour, nil
			},
		},
		"total": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewTimeBucket)
				return source.Total, nil
			},
		},
		"blocked": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewTimeBucket)
				return source.Blocked, nil
			},
		},
		"flagged": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewTimeBucket)
				return source.Flagged, nil
			},
		},
	},
})

var fraudProtectionOverviewTopSourceIPType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FraudProtectionOverviewTopSourceIP",
	Fields: graphql.Fields{
		"ipAddress": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.IPAddress, nil
			},
		},
		"geoCountryCode": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.GeoCountryCode, nil
			},
		},
		"total": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.Total, nil
			},
		},
		"blocked": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.Blocked, nil
			},
		},
		"flagged": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.Flagged, nil
			},
		},
	},
})

var fraudProtectionOverviewIPLocationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FraudProtectionOverviewIPLocation",
	Fields: graphql.Fields{
		"geoCountryCode": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewIPLocation)
				return source.GeoCountryCode, nil
			},
		},
		"total": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewIPLocation)
				return source.Total, nil
			},
		},
		"blocked": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewIPLocation)
				return source.Blocked, nil
			},
		},
		"flagged": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewIPLocation)
				return source.Flagged, nil
			},
		},
	},
})

var fraudProtectionOverviewSMSOriginType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FraudProtectionOverviewSMSOrigin",
	Fields: graphql.Fields{
		"phoneCountryCode": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSMSOrigin)
				return source.PhoneCountryCode, nil
			},
		},
		"total": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSMSOrigin)
				return source.Total, nil
			},
		},
		"blocked": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSMSOrigin)
				return source.Blocked, nil
			},
		},
		"flagged": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSMSOrigin)
				return source.Flagged, nil
			},
		},
	},
})

var fraudProtectionOverviewSendSMSType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FraudProtectionOverviewSendSMS",
	Fields: graphql.Fields{
		"total": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.TotalActions, nil
			},
		},
		"blocked": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.BlockedActions, nil
			},
		},
		"flagged": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.WarnedActions, nil
			},
		},
		"topSourceIPs": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(fraudProtectionOverviewTopSourceIPType))),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.TopSourceIPs, nil
			},
		},
		"topIPLocations": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(fraudProtectionOverviewIPLocationType))),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.TopIPLocations, nil
			},
		},
		"topSMSOrigins": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(fraudProtectionOverviewSMSOriginType))),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.TopSMSOrigins, nil
			},
		},
		"timeBuckets": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(fraudProtectionOverviewTimeBucketType))),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.TimeBuckets, nil
			},
		},
	},
})

var fraudProtectionOverviewType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FraudProtectionOverview",
	Fields: graphql.Fields{
		"sendSMS": &graphql.Field{
			Type: graphql.NewNonNull(fraudProtectionOverviewSendSMSType),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				source := p.Source.(*audit.FraudProtectionOverview)
				return source.SendSMS, nil
			},
		},
	},
})
