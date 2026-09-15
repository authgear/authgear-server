package graphqlutil

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func TestCountTopLevelMutationFields(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		operationName string
		wantCount     int
		wantErr       apierrors.Name
	}{
		{
			name: "counts mutation fields",
			query: `mutation {
				a1: createUser(input: {}) { user { id } }
				a2: createUser(input: {}) { user { id } }
				a3: createUser(input: {}) { user { id } }
				a4: createUser(input: {}) { user { id } }
				a5: createUser(input: {}) { user { id } }
			}`,
			wantCount: 5,
		},
		{
			name: "counts aliased repeats of one field separately",
			query: `mutation {
				a1: createGroup(input: {}) { group { id } }
				a2: createGroup(input: {}) { group { id } }
				a3: createGroup(input: {}) { group { id } }
			}`,
			wantCount: 3,
		},
		{
			name: "counts fragment spreads",
			query: `mutation BulkCreate {
				...CreateUsers
			}

			fragment CreateUsers on Mutation {
				a1: createUser(input: {}) { user { id } }
				a2: createUser(input: {}) { user { id } }
				a3: createUser(input: {}) { user { id } }
				a4: createUser(input: {}) { user { id } }
				a5: createUser(input: {}) { user { id } }
				a6: createUser(input: {}) { user { id } }
			}`,
			operationName: "BulkCreate",
			wantCount:     6,
		},
		{
			name: "counts inline fragments",
			query: `mutation {
				... on Mutation {
					a1: createUser(input: {}) { user { id } }
					a2: createUser(input: {}) { user { id } }
					a3: createUser(input: {}) { user { id } }
					a4: createUser(input: {}) { user { id } }
					a5: createUser(input: {}) { user { id } }
					a6: createUser(input: {}) { user { id } }
				}
			}`,
			wantCount: 6,
		},
		{
			name: "rejects fragment cycle",
			query: `mutation {
				...FragA
			}

			fragment FragA on Mutation {
				...FragB
			}

			fragment FragB on Mutation {
				...FragA
			}`,
			wantErr: apierrors.BadRequest,
		},
		{
			name: "ignores queries",
			query: `query {
				viewer { id }
			}`,
			wantCount: 0,
		},
		{
			// Several operations with no operationName is ambiguous, so no
			// operation is selected and nothing is counted. TestContextHandler
			// pins the other half of this: such a document does not execute.
			name: "returns zero for an ambiguous multi-operation document",
			query: `mutation A {
				createUser(input: {}) { user { id } }
			}

			mutation B {
				createGroup(input: {}) { group { id } }
			}`,
			wantCount: 0,
		},
		{
			name: "counts only the named operation",
			query: `mutation A {
				a1: createUser(input: {}) { user { id } }
			}

			mutation B {
				b1: createGroup(input: {}) { group { id } }
				b2: createGroup(input: {}) { group { id } }
			}`,
			operationName: "B",
			wantCount:     2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			count, err := CountTopLevelMutationFields(tc.query, tc.operationName)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if count != tc.wantCount {
					t.Fatalf("expected count %d, got %d", tc.wantCount, count)
				}
				return
			}
			apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
			if apiErr == nil || apiErr.Kind.Name != tc.wantErr {
				t.Fatalf("expected API error name %v, got %v", tc.wantErr, err)
			}
		})
	}
}

// newTestSchema builds the smallest schema that can observe whether a mutation
// actually executed: touch increments the counter it closes over.
func newTestSchema(t *testing.T, touched *int) *graphql.Schema {
	t.Helper()
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name: "Query",
			Fields: graphql.Fields{
				"ping": &graphql.Field{
					Type: graphql.String,
					Resolve: func(p graphql.ResolveParams) (any, error) {
						return "pong", nil
					},
				},
			},
		}),
		Mutation: graphql.NewObject(graphql.ObjectConfig{
			Name: "Mutation",
			Fields: graphql.Fields{
				"touch": &graphql.Field{
					Type: graphql.String,
					Resolve: func(p graphql.ResolveParams) (any, error) {
						*touched++
						return "ok", nil
					},
				},
			},
		}),
	})
	if err != nil {
		t.Fatalf("failed to build test schema: %v", err)
	}
	return &schema
}

func postGraphQL(t *testing.T, h *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	r, err := http.NewRequest("POST", "/graphql", strings.NewReader(body))
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	r.Header.Set("Content-Type", ContentTypeJSON)
	w := httptest.NewRecorder()
	h.ContextHandler(context.Background(), w, r)
	return w
}

func TestContextHandler(t *testing.T) {
	mutationBody := func(n int) string {
		var b strings.Builder
		b.WriteString("mutation {")
		for i := range n {
			b.WriteString(" a")
			b.WriteString(string(rune('0' + i)))
			b.WriteString(": touch")
		}
		b.WriteString(" }")
		payload, err := json.Marshal(map[string]any{"query": b.String()})
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}
		return string(payload)
	}

	t.Run("rejects a document over the mutation field limit without executing it", func(t *testing.T) {
		touched := 0
		h := &Handler{Schema: newTestSchema(t, &touched)}

		w := postGraphQL(t, h, mutationBody(maxMutationFieldsPerRequest+1))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
		if touched != 0 {
			t.Fatalf("expected no mutation to execute, got %d", touched)
		}
	})

	t.Run("passes the mutation field count to BeforeExecuteFn", func(t *testing.T) {
		touched := 0
		got := -1
		h := &Handler{
			Schema: newTestSchema(t, &touched),
			BeforeExecuteFn: func(ctx context.Context, mutationFieldCount int) error {
				got = mutationFieldCount
				return nil
			},
		}

		w := postGraphQL(t, h, mutationBody(3))

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if got != 3 {
			t.Fatalf("expected BeforeExecuteFn to receive 3, got %d", got)
		}
		if touched != 3 {
			t.Fatalf("expected 3 mutations to execute, got %d", touched)
		}
	})

	t.Run("passes zero to BeforeExecuteFn for a query", func(t *testing.T) {
		touched := 0
		got := -1
		h := &Handler{
			Schema: newTestSchema(t, &touched),
			BeforeExecuteFn: func(ctx context.Context, mutationFieldCount int) error {
				got = mutationFieldCount
				return nil
			},
		}

		w := postGraphQL(t, h, `{"query": "query { ping }"}`)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		if got != 0 {
			t.Fatalf("expected BeforeExecuteFn to receive 0, got %d", got)
		}
	})

	t.Run("an error from BeforeExecuteFn rejects the request without executing it", func(t *testing.T) {
		touched := 0
		h := &Handler{
			Schema: newTestSchema(t, &touched),
			BeforeExecuteFn: func(ctx context.Context, mutationFieldCount int) error {
				return apierrors.TooManyRequest.WithReason("RateLimited").New("request rate limited")
			},
		}

		w := postGraphQL(t, h, mutationBody(1))

		if w.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429, got %d", w.Code)
		}
		if touched != 0 {
			t.Fatalf("expected no mutation to execute, got %d", touched)
		}
	})

	t.Run("an ambiguous multi-operation document executes nothing", func(t *testing.T) {
		// CountTopLevelMutationFields returns 0 for this document, so
		// BeforeExecuteFn is called with 0 and any rate limit is skipped. That
		// is only safe because the document does not execute either — which is
		// what this asserts, rather than assuming it of graphql.Do.
		touched := 0
		got := -1
		h := &Handler{
			Schema: newTestSchema(t, &touched),
			BeforeExecuteFn: func(ctx context.Context, mutationFieldCount int) error {
				got = mutationFieldCount
				return nil
			},
		}

		w := postGraphQL(t, h, `{"query": "mutation A { a: touch } mutation B { b: touch }"}`)

		if got != 0 {
			t.Fatalf("expected BeforeExecuteFn to receive 0, got %d", got)
		}
		if touched != 0 {
			t.Fatalf("expected no mutation to execute, got %d", touched)
		}
		var resp struct {
			Errors []any `json:"errors"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(resp.Errors) == 0 {
			t.Fatalf("expected a GraphQL error, got %s", w.Body.String())
		}
	})
}
