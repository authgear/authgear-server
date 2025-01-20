package relay_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/location"
	"github.com/graphql-go/graphql/testutil"
	"golang.org/x/net/context"

	"github.com/authgear/authgear-server/pkg/graphqlgo/relay"
)

type user struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type photo struct {
	ID    int `json:"id"`
	Width int `json:"width"`
}

var nodeTestUserData = map[string]*user{
	"1": &user{1, "John Doe"},
	"2": &user{2, "Jane Smith"},
}
var nodeTestPhotoData = map[string]*photo{
	"3": &photo{3, 300},
	"4": &photo{4, 400},
}

// declare types first, define later in init()
// because they all depend on nodeTestDef
var nodeTestUserType *graphql.Object
var nodeTestPhotoType *graphql.Object

var nodeTestDef = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
	IDFetcher: func(id string, info graphql.ResolveInfo, ctx context.Context) (interface{}, error) {
		if user, ok := nodeTestUserData[id]; ok {
			return user, nil
		}
		if photo, ok := nodeTestPhotoData[id]; ok {
			return photo, nil
		}
		return nil, errors.New("Unknown node")
	},
	TypeResolve: func(p graphql.ResolveTypeParams) *graphql.Object {
		switch p.Value.(type) {
		case *user:
			return nodeTestUserType
		case *photo:
			return nodeTestPhotoType
		default:
			panic(fmt.Sprintf("Unknown object type `%v`", p.Value))
		}
	},
})
var nodeTestQueryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"node": nodeTestDef.NodeField,
	},
})

// be careful not to define schema here, since nodeTestUserType and nodeTestPhotoType wouldn't be defined till init()
var nodeTestSchema graphql.Schema

func init() {
	nodeTestUserType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
		},
		Interfaces: []*graphql.Interface{nodeTestDef.NodeInterface},
	})
	nodeTestPhotoType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Photo",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.ID),
			},
			"width": &graphql.Field{
				Type: graphql.Int,
			},
		},
		Interfaces: []*graphql.Interface{nodeTestDef.NodeInterface},
	})

	nodeTestSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query: nodeTestQueryType,
		Types: []graphql.Type{nodeTestUserType, nodeTestPhotoType},
	})
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectIDForUsers(t *testing.T) {
	query := `{
        node(id: "1") {
          id
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id": "1",
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectIDForPhotos(t *testing.T) {
	query := `{
        node(id: "4") {
          id
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id": "4",
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectNameForUsers(t *testing.T) {
	query := `{
        node(id: "1") {
          id
          ... on User {
            name
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id":   "1",
				"name": "John Doe",
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectWidthForPhotos(t *testing.T) {
	query := `{
        node(id: "4") {
          id
          ... on Photo {
            width
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id":    "4",
				"width": 400,
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectTypeNameForUsers(t *testing.T) {
	query := `{
        node(id: "1") {
          id
          __typename
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id":         "1",
				"__typename": "User",
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_GetsTheCorrectTypeNameForPhotos(t *testing.T) {
	query := `{
        node(id: "4") {
          id
          __typename
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id":         "4",
				"__typename": "Photo",
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_AllowsRefetching_IgnoresPhotoFragmentsOnUser(t *testing.T) {
	query := `{
        node(id: "1") {
          id
          ... on Photo {
            width
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": map[string]interface{}{
				"id": "1",
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

// This test is skipped because we cannot simply use reflect.DeepEqual to match the exact result.
func SkipTestNodeInterfaceAndFields_AllowsRefetching_ReturnsNullForBadIDs(t *testing.T) {
	query := `{
        node(id: "5") {
          id
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"node": nil,
		},
		Errors: []gqlerrors.FormattedError{
			{
				Message:   "Unknown node",
				Locations: []location.SourceLocation{},
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_CorrectlyIntrospects_HasCorrectNodeInterface(t *testing.T) {
	query := `{
        __type(name: "Node") {
          name
          kind
          fields {
            name
            type {
              kind
              ofType {
                name
                kind
              }
            }
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"__type": map[string]interface{}{
				"name": "Node",
				"kind": "INTERFACE",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "id",
						"type": map[string]interface{}{
							"kind": "NON_NULL",
							"ofType": map[string]interface{}{
								"name": "ID",
								"kind": "SCALAR",
							},
						},
					},
				},
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestNodeInterfaceAndFields_CorrectlyIntrospects_HasCorrectNodeRootField(t *testing.T) {
	query := `{
        __schema {
          queryType {
            fields {
              name
              type {
                name
                kind
              }
              args {
                name
                type {
                  kind
                  ofType {
                    name
                    kind
                  }
                }
              }
            }
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"name": "node",
							"type": map[string]interface{}{
								"name": "Node",
								"kind": "INTERFACE",
							},
							"args": []interface{}{
								map[string]interface{}{
									"name": "id",
									"type": map[string]interface{}{
										"kind": "NON_NULL",
										"ofType": map[string]interface{}{
											"name": "ID",
											"kind": "SCALAR",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        nodeTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
