package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/lib/audit"
)

var fraudProtectionOverviewTopSourceIPType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FraudProtectionOverviewTopSourceIP",
	Fields: graphql.Fields{
		"ipAddress": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.IPAddress, nil
			},
		},
		"total": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.Total, nil
			},
		},
		"blocked": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.Blocked, nil
			},
		},
		"flagged": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
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
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.TotalActions, nil
			},
		},
		"blocked": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.BlockedActions, nil
			},
		},
		"flagged": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.WarnedActions, nil
			},
		},
		"topSourceIPs": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(fraudProtectionOverviewTopSourceIPType))),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewSendSMS)
				return source.TopSourceIPs, nil
			},
		},
	},
})

var fraudProtectionOverviewType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FraudProtectionOverview",
	Fields: graphql.Fields{
		"sendSMS": &graphql.Field{
			Type: graphql.NewNonNull(fraudProtectionOverviewSendSMSType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*audit.FraudProtectionOverview)
				return source.SendSMS, nil
			},
		},
	},
})
