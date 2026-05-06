package relay_test

import (
	context0 "context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/location"
	"github.com/graphql-go/graphql/testutil"

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
	MutateAndGetPayload: func(inputMap map[string]any, info graphql.ResolveInfo, ctx context0.Context) (map[string]any, error) {
		return map[string]any{
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
	MutateAndGetPayload: func(inputMap map[string]any, info graphql.ResolveInfo, ctx context0.Context) (map[string]any, error) {
		return map[string]any(nil), NotFoundError
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
	MutateAndGetPayload: func(inputMap map[string]any, info graphql.ResolveInfo, ctx context0.Context) (map[string]any, error) {
		c := make(chan int)
		go testAsyncDataMutation(&c)
		result := <-c
		return map[string]any{
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
		Data: map[string]any{
			"simpleMutation": map[string]any{
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
		Data: map[string]any{
			"simplePromiseMutation": map[string]any{
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
		Data: map[string]any{
			"__type": map[string]any{
				"name": "SimpleMutationInput",
				"kind": "INPUT_OBJECT",
				"inputFields": []any{
					map[string]any{
						"name": "clientMutationId",
						"type": map[string]any{
							"name": nil,
							"kind": "NON_NULL",
							"ofType": map[string]any{
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
		Data: map[string]any{
			"__type": map[string]any{
				"name": "SimpleMutationPayload",
				"kind": "OBJECT",
				"fields": []any{
					map[string]any{
						"name": "result",
						"type": map[string]any{
							"name":   "Int",
							"kind":   "SCALAR",
							"ofType": nil,
						},
					},
					map[string]any{
						"name": "clientMutationId",
						"type": map[string]any{
							"name": nil,
							"kind": "NON_NULL",
							"ofType": map[string]any{
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
	if !testutil.ContainSubset(result.Data.(map[string]any), expected.Data.(map[string]any)) {
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
		Data: map[string]any{
			"__schema": map[string]any{
				"mutationType": map[string]any{
					"fields": []any{
						map[string]any{
							"name": "simpleMutation",
							"args": []any{
								map[string]any{
									"name": "input",
									"type": map[string]any{
										"name": nil,
										"kind": "NON_NULL",
										"ofType": map[string]any{
											"name": "SimpleMutationInput",
											"kind": "INPUT_OBJECT",
										},
									},
								},
							},
							"type": map[string]any{
								"name": "SimpleMutationPayload",
								"kind": "OBJECT",
							},
						},
						map[string]any{
							"name": "simplePromiseMutation",
							"args": []any{
								map[string]any{
									"name": "input",
									"type": map[string]any{
										"name": nil,
										"kind": "NON_NULL",
										"ofType": map[string]any{
											"name": "SimplePromiseMutationInput",
											"kind": "INPUT_OBJECT",
										},
									},
								},
							},
							"type": map[string]any{
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
	if !testutil.ContainSubset(result.Data.(map[string]any), expected.Data.(map[string]any)) {
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
		Data: map[string]any{
			"simpleMutation": any(nil),
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
