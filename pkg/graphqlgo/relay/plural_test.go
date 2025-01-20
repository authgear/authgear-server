package relay_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/location"
	"github.com/graphql-go/graphql/testutil"
	"github.com/kr/pretty"

	"github.com/authgear/authgear-server/pkg/graphqlgo/relay"
)

var pluralTestUserType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"username": &graphql.Field{
			Type: graphql.String,
		},
		"url": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var pluralTestQueryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"usernames": relay.PluralIdentifyingRootField(relay.PluralIdentifyingRootFieldConfig{
			ArgName:     "usernames",
			Description: "Map from a username to the user",
			InputType:   graphql.String,
			OutputType:  pluralTestUserType,
			ResolveSingleInput: func(username interface{}) interface{} {
				return map[string]interface{}{
					"username": fmt.Sprintf("%v", username),
					"url":      fmt.Sprintf("www.facebook.com/%v", username),
				}
			},
		}),
	},
})

var pluralTestSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query: pluralTestQueryType,
})

func TestPluralIdentifyingRootField_AllowsFetching(t *testing.T) {
	query := `{
      usernames(usernames:["dschafer", "leebyron", "schrockn"]) {
        username
        url
      }
    }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"usernames": []interface{}{
				map[string]interface{}{
					"username": "dschafer",
					"url":      "www.facebook.com/dschafer",
				},
				map[string]interface{}{
					"username": "leebyron",
					"url":      "www.facebook.com/leebyron",
				},
				map[string]interface{}{
					"username": "schrockn",
					"url":      "www.facebook.com/schrockn",
				},
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        pluralTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestPluralIdentifyingRootField_CorrectlyIntrospects(t *testing.T) {
	query := `{
      __schema {
        queryType {
          fields {
            name
            args {
              name
              type {
                kind
                ofType {
                  kind
                  ofType {
                    kind
                    ofType {
                      name
                      kind
                    }
                  }
                }
              }
            }
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
    }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"name": "usernames",
							"args": []interface{}{
								map[string]interface{}{
									"name": "usernames",
									"type": map[string]interface{}{
										"kind": "NON_NULL",
										"ofType": map[string]interface{}{
											"kind": "LIST",
											"ofType": map[string]interface{}{
												"kind": "NON_NULL",
												"ofType": map[string]interface{}{
													"name": "String",
													"kind": "SCALAR",
												},
											},
										},
									},
								},
							},
							"type": map[string]interface{}{
								"kind": "LIST",
								"ofType": map[string]interface{}{
									"name": "User",
									"kind": "OBJECT",
								},
							},
						},
					},
				},
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        pluralTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestPluralIdentifyingRootField_Configuration_ResolveSingleInputIsNil(t *testing.T) {

	var pluralTestQueryType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"usernames": relay.PluralIdentifyingRootField(relay.PluralIdentifyingRootFieldConfig{
				ArgName:     "usernames",
				Description: "Map from a username to the user",
				InputType:   graphql.String,
				OutputType:  pluralTestUserType,
			}),
		},
	})

	var pluralTestSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query: pluralTestQueryType,
	})

	query := `{
      usernames(usernames:["dschafer", "leebyron", "schrockn"]) {
        username
        url
      }
    }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"usernames": nil,
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        pluralTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestPluralIdentifyingRootField_Configuration_ArgNames_WrongArgNameSpecified(t *testing.T) {

	t.Skipf("Pending `validator` implementation")
	query := `{
      usernames(usernamesMisspelled:["dschafer", "leebyron", "schrockn"]) {
        username
        url
      }
    }`
	expected := &graphql.Result{
		Data: nil,
		Errors: []gqlerrors.FormattedError{
			gqlerrors.FormattedError{
				Message: `Unknown argument "usernamesMisspelled" on field "usernames" of type "Query".`,
				Locations: []location.SourceLocation{
					location.SourceLocation{Line: 2, Column: 17},
				},
			},
			gqlerrors.FormattedError{
				Message: `Field "usernames" argument "usernames" of type "[String!]!" is required but not provided.`,
				Locations: []location.SourceLocation{
					location.SourceLocation{Line: 2, Column: 7},
				},
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        pluralTestSchema,
		RequestString: query,
	})
	pretty.Println(result)
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
