package graphql

import (
	"time"

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
		"totalActions": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.TotalActions, nil
			},
		},
		"blockedActions": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.BlockedActions, nil
			},
		},
		"warnedActions": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(audit.FraudProtectionOverviewIP)
				return source.WarnedActions, nil
			},
		},
	},
})

var fraudProtectionOverviewType = graphql.NewObject(graphql.ObjectConfig{
	Name: "FraudProtectionOverview",
	Fields: graphql.Fields{
		"totalActions": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*audit.FraudProtectionOverview)
				return source.TotalActions, nil
			},
		},
		"allowedActions": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*audit.FraudProtectionOverview)
				return source.AllowedActions, nil
			},
		},
		"blockedActions": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*audit.FraudProtectionOverview)
				return source.BlockedActions, nil
			},
		},
		"warnedActions": &graphql.Field{
			Type: graphql.NewNonNull(graphql.Int),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*audit.FraudProtectionOverview)
				return source.WarnedActions, nil
			},
		},
		"topSourceIPs": &graphql.Field{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(fraudProtectionOverviewTopSourceIPType))),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*audit.FraudProtectionOverview)
				return source.TopSourceIPs, nil
			},
		},
	},
})

var fraudProtectionOverviewArgs = graphql.FieldConfigArgument{
	"rangeFrom": &graphql.ArgumentConfig{
		Type: graphql.DateTime,
	},
	"rangeTo": &graphql.ArgumentConfig{
		Type: graphql.DateTime,
	},
}

func fraudProtectionOverviewResolveOpts(p graphql.ResolveParams) audit.QueryPageOptions {
	var rangeFrom *time.Time
	if t, ok := p.Args["rangeFrom"].(time.Time); ok {
		rangeFrom = &t
	}

	var rangeTo *time.Time
	if t, ok := p.Args["rangeTo"].(time.Time); ok {
		rangeTo = &t
	}

	return audit.QueryPageOptions{
		RangeFrom: rangeFrom,
		RangeTo:   rangeTo,
	}
}
