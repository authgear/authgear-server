package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
)

var collaboratorRole = graphql.NewEnum(graphql.EnumConfig{
	Name: "CollaboratorRole",
	Values: graphql.EnumValueConfigMap{
		"OWNER": &graphql.EnumValueConfig{
			Value: model.CollaboratorRoleOwner,
		},
		"EDITOR": &graphql.EnumValueConfig{
			Value: model.CollaboratorRoleEditor,
		},
	},
})

var collaborator = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Collaborator",
	Description: "Collaborator of an app",
	Fields: graphql.Fields{
		"id": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},

		// AppID is intentionally excluded because
		// if we wanted to reference back the app, we would add "app"
		// But adding it would result in a circular schema.

		"user": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*model.Collaborator)
				gqlCtx := GQLContext(p.Context)
				return gqlCtx.Users.Load(source.UserID).Value, nil
			},
		},

		"createdAt": &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
		"role":      &graphql.Field{Type: graphql.NewNonNull(collaboratorRole)},
	},
})

var collaboratorInvitation = graphql.NewObject(graphql.ObjectConfig{
	Name:        "CollaboratorInvitation",
	Description: "Collaborator invitation of an app",
	Fields: graphql.Fields{
		"id": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},

		// AppID is intentionally excluded because
		// if we wanted to reference back the app, we would add "app"
		// But adding it would result in a circular schema.

		"invitedBy": &graphql.Field{
			Type: graphql.NewNonNull(nodeUser),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				source := p.Source.(*model.CollaboratorInvitation)
				gqlCtx := GQLContext(p.Context)
				return gqlCtx.Users.Load(source.InvitedBy).Value, nil
			},
		},

		"inviteeEmail": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"createdAt":    &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
		"expireAt":     &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
	},
})
