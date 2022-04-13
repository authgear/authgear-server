package graphql

import (
	"github.com/graphql-go/graphql"
)

var tutorialStatus = graphql.NewObject(graphql.ObjectConfig{
	Name:        "TutorialStatus",
	Description: "Tutorial status of an app",
	Fields: graphql.Fields{
		"appID": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"data":  &graphql.Field{Type: graphql.NewNonNull(TutorialStatusData)},
	},
})
