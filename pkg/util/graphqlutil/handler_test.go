package graphqlutil

import (
	"context"
	"testing"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func TestValidateMutationFieldCount(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		operationName string
		limit         int
		wantErr       apierrors.Name
	}{
		{
			name: "allows mutation at limit",
			query: `mutation {
				a1: createUser(input: {}) { user { id } }
				a2: createUser(input: {}) { user { id } }
				a3: createUser(input: {}) { user { id } }
				a4: createUser(input: {}) { user { id } }
				a5: createUser(input: {}) { user { id } }
			}`,
			limit: 5,
		},
		{
			name: "rejects mutation over limit",
			query: `mutation {
				a1: createUser(input: {}) { user { id } }
				a2: createUser(input: {}) { user { id } }
				a3: createUser(input: {}) { user { id } }
				a4: createUser(input: {}) { user { id } }
				a5: createUser(input: {}) { user { id } }
				a6: createUser(input: {}) { user { id } }
			}`,
			limit:   5,
			wantErr: apierrors.BadRequest,
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
			limit:         5,
			wantErr:       apierrors.BadRequest,
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
			limit:   5,
			wantErr: apierrors.BadRequest,
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
			limit:   5,
			wantErr: apierrors.BadRequest,
		},
		{
			name: "ignores queries",
			query: `query {
				viewer { id }
			}`,
			limit: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMutationFieldCount(tc.query, tc.operationName, tc.limit)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tc.wantErr != "" {
				apiErr := apierrors.AsAPIErrorWithContext(context.Background(), err)
				if apiErr == nil || apiErr.Kind.Name != tc.wantErr {
					t.Fatalf("expected API error name %v, got %v", tc.wantErr, err)
				}
			}
		})
	}
}
