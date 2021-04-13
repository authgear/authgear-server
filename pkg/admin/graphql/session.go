package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

const typeSession = "Session"

var sessionType = graphql.NewEnum(graphql.EnumConfig{
	Name: "SessionType",
	Values: graphql.EnumValueConfigMap{
		"IDP": &graphql.EnumValueConfig{
			Value: model.SessionTypeIDP,
		},
		"OFFLINE_GRANT": &graphql.EnumValueConfig{
			Value: model.SessionTypeOfflineGrant,
		},
	},
})

var nodeSession = entity(
	graphql.NewObject(graphql.ObjectConfig{
		Name: typeSession,
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface,
			entityInterface,
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeSession, nil),
			"createdAt": entityCreatedAtField(nil),
			"updatedAt": entityUpdatedAtField(nil),
			"type": &graphql.Field{
				Type: graphql.NewNonNull(sessionType),
			},
			"acr": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"amr": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
			},
			"lastAccessedAt": &graphql.Field{
				Type: graphql.NewNonNull(graphql.DateTime),
			},
			"lastAccessedByIP": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"createdByIP": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"displayName": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	}),
	&model.Session{},
	func(ctx *Context, id string) (interface{}, error) {
		s, err := ctx.SessionFacade.Get(id)
		if err != nil {
			return nil, err
		}
		return s.ToAPIModel(), nil
	},
)

var connSession = graphqlutil.NewConnectionDef(nodeSession)
