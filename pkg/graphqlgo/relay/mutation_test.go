package relay_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/location"
	"github.com/graphql-go/graphql/testutil"
	"golang.org/x/net/context"

	"github.com/authgear/authgear-server/pkg/graphqlgo/relay"
)

func testAsyncDataMutation(resultChan *chan int) {
	// simulate async data mutation
	time.Sleep(time.Second * 1)
	*resultChan <- int(1)
}

var simpleMutationTest = relay.MutationWithClientMutationID(relay.MutationConfig{
	Name:        "SimpleMutation",
	InputFields: graphql.InputObjectConfigFieldMap{},
	OutputFields: graphql.Fields{
		"result": &graphql.Field{
			Type: graphql.Int,
		},
	},
	MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
		return map[string]interface{}{
			"result": 1,
		}, nil
	},
})

var NotFoundError = errors.New("not found")

var simpleMutationErrorTest = relay.MutationWithClientMutationID(relay.MutationConfig{
	Name:        "SimpleMutation",
	InputFields: graphql.InputObjectConfigFieldMap{},
	OutputFields: graphql.Fields{
		"result": &graphql.Field{
			Type: graphql.Int,
		},
	},
	MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
		return map[string]interface{}(nil), NotFoundError
	},
})

// async mutation
var simplePromiseMutationTest = relay.MutationWithClientMutationID(relay.MutationConfig{
	Name:        "SimplePromiseMutation",
	InputFields: graphql.InputObjectConfigFieldMap{},
	OutputFields: graphql.Fields{
		"result": &graphql.Field{
			Type: graphql.Int,
		},
	},
	MutateAndGetPayload: func(inputMap map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
		c := make(chan int)
		go testAsyncDataMutation(&c)
		result := <-c
		return map[string]interface{}{
			"result": result,
		}, nil
	},
})

var mutationTestType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"simpleMutation":        simpleMutationTest,
		"simplePromiseMutation": simplePromiseMutationTest,
	},
})

var mutationTestTypeError = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"simpleMutation":        simpleMutationErrorTest,
		"simplePromiseMutation": simplePromiseMutationTest,
	},
})

var mutationTestSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    mutationTestType,
	Mutation: mutationTestType,
})

var mutationTestSchemaError, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    mutationTestType,
	Mutation: mutationTestTypeError,
})

func TestMutation_WithClientMutationId_BehavesCorrectly_RequiresAnArgument(t *testing.T) {
	t.Skipf("Pending `validator` implementation")
	query := `
        mutation M {
          simpleMutation {
            result
          }
        }
      `
	expected := &graphql.Result{
		Errors: []gqlerrors.FormattedError{
			gqlerrors.FormattedError{
				Message: `Field "simpleMutation" argument "input" of type "SimpleMutationInput!" is required but not provided.`,
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        mutationTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestMutation_WithClientMutationId_BehavesCorrectly_ReturnsTheSameClientMutationId(t *testing.T) {
	query := `
        mutation M {
          simpleMutation(input: {clientMutationId: "abc"}) {
            result
            clientMutationId
          }
        }
      `
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"simpleMutation": map[string]interface{}{
				"result":           1,
				"clientMutationId": "abc",
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        mutationTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

// Async mutation using channels
func TestMutation_WithClientMutationId_BehavesCorrectly_SupportsPromiseMutations(t *testing.T) {
	query := `
        mutation M {
          simplePromiseMutation(input: {clientMutationId: "abc"}) {
            result
            clientMutationId
          }
        }
      `
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"simplePromiseMutation": map[string]interface{}{
				"result":           1,
				"clientMutationId": "abc",
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        mutationTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestMutation_IntrospectsCorrectly_ContainsCorrectInput(t *testing.T) {
	query := `{
        __type(name: "SimpleMutationInput") {
          name
          kind
          inputFields {
            name
            type {
              name
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
				"name": "SimpleMutationInput",
				"kind": "INPUT_OBJECT",
				"inputFields": []interface{}{
					map[string]interface{}{
						"name": "clientMutationId",
						"type": map[string]interface{}{
							"name": nil,
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
	}
	result := graphql.Do(graphql.Params{
		Schema:        mutationTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
func TestMutation_IntrospectsCorrectly_ContainsCorrectPayload(t *testing.T) {
	query := `{
        __type(name: "SimpleMutationPayload") {
          name
          kind
          fields {
            name
            type {
              name
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
				"name": "SimpleMutationPayload",
				"kind": "OBJECT",
				"fields": []interface{}{
					map[string]interface{}{
						"name": "result",
						"type": map[string]interface{}{
							"name":   "Int",
							"kind":   "SCALAR",
							"ofType": nil,
						},
					},
					map[string]interface{}{
						"name": "clientMutationId",
						"type": map[string]interface{}{
							"name": nil,
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
	}
	result := graphql.Do(graphql.Params{
		Schema:        mutationTestSchema,
		RequestString: query,
	})
	if !testutil.ContainSubset(result.Data.(map[string]interface{}), expected.Data.(map[string]interface{})) {
		t.Fatalf("unexpected, result does not contain subset of expected data")
	}
}
func TestMutation_IntrospectsCorrectly_ContainsCorrectField(t *testing.T) {
	query := `{
        __schema {
          mutationType {
            fields {
              name
              args {
                name
                type {
                  name
                  kind
                  ofType {
                    name
                    kind
                  }
                }
              }
              type {
                name
                kind
              }
            }
          }
        }
      }`
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"__schema": map[string]interface{}{
				"mutationType": map[string]interface{}{
					"fields": []interface{}{
						map[string]interface{}{
							"name": "simpleMutation",
							"args": []interface{}{
								map[string]interface{}{
									"name": "input",
									"type": map[string]interface{}{
										"name": nil,
										"kind": "NON_NULL",
										"ofType": map[string]interface{}{
											"name": "SimpleMutationInput",
											"kind": "INPUT_OBJECT",
										},
									},
								},
							},
							"type": map[string]interface{}{
								"name": "SimpleMutationPayload",
								"kind": "OBJECT",
							},
						},
						map[string]interface{}{
							"name": "simplePromiseMutation",
							"args": []interface{}{
								map[string]interface{}{
									"name": "input",
									"type": map[string]interface{}{
										"name": nil,
										"kind": "NON_NULL",
										"ofType": map[string]interface{}{
											"name": "SimplePromiseMutationInput",
											"kind": "INPUT_OBJECT",
										},
									},
								},
							},
							"type": map[string]interface{}{
								"name": "SimplePromiseMutationPayload",
								"kind": "OBJECT",
							},
						},
					},
				},
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        mutationTestSchema,
		RequestString: query,
	})
	if !testutil.ContainSubset(result.Data.(map[string]interface{}), expected.Data.(map[string]interface{})) {
		t.Fatalf("unexpected, result does not contain subset of expected data")
	}
}

// This test is skipped because we cannot simply use reflect.DeepEqual to match the exact result.
func SkipTestMutateAndGetPayload_AddsErrors(t *testing.T) {
	query := `
        mutation M {
          simpleMutation(input: {clientMutationId: "abc"}) {
            result
            clientMutationId
          }
        }
      `
	expected := &graphql.Result{
		Data: map[string]interface{}{
			"simpleMutation": interface{}(nil),
		},
		Errors: []gqlerrors.FormattedError{
			gqlerrors.FormattedError{
				Message:   NotFoundError.Error(),
				Locations: []location.SourceLocation{},
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        mutationTestSchemaError,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
