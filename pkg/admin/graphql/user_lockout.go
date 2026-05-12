package graphql

import (
	"github.com/graphql-go/graphql"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
)

var lockedIPType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "LockedIP",
	Description: "A locked IP address and when its lock expires",
	Fields: graphql.Fields{
		"ipAddress": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The locked IP address",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(apimodel.LockedIP).IPAddress, nil
			},
		},
		"lockedUntil": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.DateTime),
			Description: "The time the lock for this IP expires",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(apimodel.LockedIP).LockedUntil, nil
			},
		},
	},
})

var accountLockoutType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "AccountLockout",
	Description: "The account lockout state of a user",
	Fields: graphql.Fields{
		"lockoutType": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.String),
			Description: "The configured lockout type: \"per_user\" or \"per_user_per_ip\"",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return string(p.Source.(*apimodel.AccountLockoutStatus).LockoutType), nil
			},
		},
		"isLocked": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.Boolean),
			Description: "Whether the user is currently locked",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(*apimodel.AccountLockoutStatus).IsLocked, nil
			},
		},
		"lockedUntil": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "When the global lock expires. Non-nil only for per_user lockout type",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source.(*apimodel.AccountLockoutStatus).LockedUntil, nil
			},
		},
		"lockedIPs": &graphql.Field{
			Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(lockedIPType))),
			Description: "Locked IPs ordered by lockedUntil descending. Non-empty only for per_user_per_ip lockout type",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ips := p.Source.(*apimodel.AccountLockoutStatus).LockedIPs
				out := make([]interface{}, len(ips))
				for i, ip := range ips {
					out[i] = ip
				}
				return out, nil
			},
		},
	},
})
