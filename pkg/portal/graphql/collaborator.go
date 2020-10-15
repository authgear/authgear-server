package graphql

import (
	"github.com/graphql-go/graphql"
)

var collaborator = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Collaborator",
	Description: "Collaborator of an app",
	Fields: graphql.Fields{
		"id": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},

		// AppID is intentionally excluded because
		// if we wanted to reference back the app, we would add "app"
		// But adding it would result in a circular schema.

		// The value is a plain user ID.
		// Is there any better way to handle this field?
		"userID": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},

		"createdAt": &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
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

		// The value is a plain user ID.
		// Is there any better way to handle this field?
		"invitedBy": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},

		"inviteeEmail": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"createdAt":    &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
		"expireAt":     &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
	},
})
