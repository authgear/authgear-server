package relay_test

import (
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"

	"github.com/authgear/authgear-server/pkg/graphqlgo/relay"
)

var connectionTestAllUsers = []interface{}{
	&user{Name: "Dan"},
	&user{Name: "Nick"},
	&user{Name: "Lee"},
	&user{Name: "Joe"},
	&user{Name: "Tim"},
}
var connectionTestUserType *graphql.Object
var connectionTestQueryType *graphql.Object
var connectionTestSchema graphql.Schema
var connectionTestConnectionDef *relay.GraphQLConnectionDefinitions

func init() {
	connectionTestUserType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"name": &graphql.Field{
				Type: graphql.String,
			},
			// re-define `friends` field later because `connectionTestUserType` has `connectionTestConnectionDef` has `connectionTestUserType` (cyclic-reference)
			"friends": &graphql.Field{},
		},
	})

	connectionTestConnectionDef = relay.ConnectionDefinitions(relay.ConnectionConfig{
		Name:     "Friend",
		NodeType: connectionTestUserType,
		EdgeFields: graphql.Fields{
			"friendshipTime": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "Yesterday", nil
				},
			},
		},
		ConnectionFields: graphql.Fields{
			"totalCount": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return len(connectionTestAllUsers), nil
				},
			},
		},
	})

	// define `friends` field here after getting connection definition
	connectionTestUserType.AddFieldConfig("friends", &graphql.Field{
		Type: connectionTestConnectionDef.ConnectionType,
		Args: relay.ConnectionArgs,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			arg := relay.NewConnectionArguments(p.Args)
			res := relay.ConnectionFromArray(connectionTestAllUsers, arg)
			return res, nil
		},
	})

	connectionTestQueryType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"user": &graphql.Field{
				Type: connectionTestUserType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return connectionTestAllUsers[0], nil
				},
			},
		},
	})
	var err error
	connectionTestSchema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: connectionTestQueryType,
	})
	if err != nil {
		panic(err)
	}

}

func TestConnectionDefinition_IncludesConnectionAndEdgeFields(t *testing.T) {
	query := `
      query FriendsQuery {
        user {
          friends(first: 2) {
            totalCount
            edges {
              friendshipTime
              node {
                name
              }
            }
          }
        }
      }
    `
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"friends": map[string]interface{}{
					"totalCount": 5,
					"edges": []interface{}{
						map[string]interface{}{
							"friendshipTime": "Yesterday",
							"node": map[string]interface{}{
								"name": "Dan",
							},
						},
						map[string]interface{}{
							"friendshipTime": "Yesterday",
							"node": map[string]interface{}{
								"name": "Nick",
							},
						},
					},
				},
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        connectionTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
